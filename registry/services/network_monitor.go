package services

import (
	"log"
	"registry/utils"
	"strconv"
	"time"
)

func checkForDeadPeers() {
	MONITOR_TIMER := utils.GetIntegerEnvironmentVariable("MONITOR_TIMER")
	HEARTBEAT_THR := utils.GetIntegerEnvironmentVariable("HEARTBEAT_THR")
	for {
		time.Sleep(time.Duration(MONITOR_TIMER) * time.Second)

		heartbeatMap.mutex.Lock()
		graphMap.mutex.Lock()
		connectionMap.mutex.Lock()
		lastCheckTime := heartbeatMap.lastChecked
		for edgePeer, lastHeartbeatTime := range heartbeatMap.heartbeats {
			if lastCheckTime.Sub(lastHeartbeatTime).Seconds() > float64(HEARTBEAT_THR) {
				//Il peer viene considerato caduto e viene rimosso dalla rete
				removeDeadNode(edgePeer)
			}
		}
		connectionMap.mutex.Unlock()
		graphMap.mutex.Unlock()
		heartbeatMap.mutex.Unlock()
	}

}

func removeDeadNode(deadPeer EdgePeer) {
	// Rimuove un nodo morto dalla rete
	// Assume che il lock sulle strutture dati sia stato preso dal chiamante
	delete(graphMap.peerMap, deadPeer)
	for otherPeer, otherPeerConnections := range graphMap.peerMap {
		var deadPeerIndex int
		for connectedPeerIndex, connectedPeer := range graphMap.peerMap[otherPeer] {
			if connectedPeer == deadPeer {
				deadPeerIndex = connectedPeerIndex
				break
			}
		}
		otherPeerConnections[deadPeerIndex] = otherPeerConnections[len(otherPeerConnections)-1]
		graphMap.peerMap[otherPeer] = otherPeerConnections[:len(otherPeerConnections)-1]
	}

	deadPeerConn := connectionMap.connections[deadPeer]
	deadPeerConn.Close()
	delete(connectionMap.connections, deadPeer)
}

// Periodically check if network is connected
func monitorNetwork() {
	monitorTimerString := utils.GetEnvironmentVariable("MONITOR_TIMER")
	MONITOR_TIMER, err := strconv.ParseInt(monitorTimerString, 10, 64)
	utils.ExitOnError("Impossibile fare il parsing di MONITOR_TIME", err)
	for {
		time.Sleep(time.Duration(MONITOR_TIMER) * time.Second)
		solveNetworkPartitions()
	}
}

// Checks for network partitions and solve them if any is found
func solveNetworkPartitions() {
	// Is a different function because network partitions solve is necessary for PeerExit and for PeerCrash too
	graphMap.mutex.Lock()
	defer graphMap.mutex.Unlock()
	connectionMap.mutex.RLock()
	defer connectionMap.mutex.RUnlock()
	connectedComponents := FindConnectedComponents(graphMap.peerMap)

	if len(connectedComponents) > 1 {
		unifyNetwork(connectedComponents)
	}

}

/*
Solves network partitions.
Network partitions are solved connecting two consecutive components in components array
*/
func unifyNetwork(connectedComponents [][]EdgePeer) {

	for componentIndex := 0; componentIndex < len(connectedComponents)-1; componentIndex++ {
		firstComponent := connectedComponents[componentIndex]
		secondComponent := connectedComponents[componentIndex+1]
		unifyTwoComponents(firstComponent, secondComponent)
	}
}

// Unifies two components of network
func unifyTwoComponents(firstComponent []EdgePeer, secondComponent []EdgePeer) {
	createdAnEdge := false
	for _, firstCompNode := range firstComponent {
		for _, secondCompNode := range secondComponent {
			var thereIsEdge bool
			if !createdAnEdge {
				// To ensure at least an edge is created between the two components
				thereIsEdge = true
			} else {
				thereIsEdge = existsEdge()
			}
			if thereIsEdge {
				// TODO Decidere se l'aggiornamento della rete viene fatto prima o dopo:
				// se fatto prima potrebbero fallire le call, se fatto dopo potrebbe essere inconsistente la rete
				graphMap.peerMap[firstCompNode] = append(graphMap.peerMap[firstCompNode], secondCompNode)
				graphMap.peerMap[secondCompNode] = append(graphMap.peerMap[secondCompNode], firstCompNode)

				firstNodeConn := connectionMap.connections[firstCompNode]
				secondNodeConn := connectionMap.connections[secondCompNode]

				err_1 := firstNodeConn.Call("PeerService.AddNeighbour", secondCompNode, nil)
				err_2 := secondNodeConn.Call("PeerService.AddNeighbour", firstCompNode, nil)
				if err_1 != nil && err_2 != nil {
					createdAnEdge = true
				}

			}
		}
	}
}

func findNeighboursForPeer(edgePeer EdgePeer) []EdgePeer {
	// peerNum := len(graphMap.peerMap)
	createdEdge := false
	neighboursList := []EdgePeer{}
	for peer := range graphMap.peerMap {
		// TODO Fare prima shuffle delle chiavi per non legare tutti i nodi al primo che viene restituito
		var thereIsEdge bool
		if !createdEdge {
			thereIsEdge = true
		} else {
			thereIsEdge = existsEdge()
		}
		if thereIsEdge {
			neighboursList = append(neighboursList, peer)
		}
	}
	log.Println(neighboursList)
	return neighboursList
}
