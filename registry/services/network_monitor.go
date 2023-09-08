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
		checkFunction(float64(HEARTBEAT_THR))
	}
}

func checkFunction(heartbeatThr float64) {
	peerMap.mutex.Lock()
	defer peerMap.mutex.Unlock()

	lastCheckTime := peerMap.heartbeatCheckTime
	for edgePeer, lastHeartbeatTime := range peerMap.heartbeats {
		if lastCheckTime.Sub(lastHeartbeatTime).Seconds() > heartbeatThr {
			//Il peer viene considerato caduto e viene rimosso dalla rete
			log.Println("Trovato Peer Morto >>> " + edgePeer.PeerAddr + "\r\n")
			deadPeerConn, isInMap := peerMap.connections[edgePeer]
			if isInMap {
				deadPeerConn.Close()
				delete(peerMap.connections, edgePeer)
			}
			delete(peerMap.heartbeats, edgePeer)
			//PrintGraph(graphMap.peerMap)
		}
	}
	peerMap.heartbeatCheckTime = time.Now()
}

func removeDeadNode(deadPeer EdgePeer) {
	// Rimuove un nodo morto dalla rete
	// Assume che il lock sulle strutture dati sia stato preso dal chiamante
	deadPeerConn, isInMap := peerMap.connections[deadPeer]
	if isInMap {
		deadPeerConn.Close()
	}

	delete(peerMap.heartbeats, deadPeer)
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
	peerMap.mutex.Lock()
	defer peerMap.mutex.Unlock()

	graph := NewGraph()

	connectedComponents := graph.FindConnectedComponents()

	if len(connectedComponents) > 1 {
		log.Printf("Trovata partizione di rete\r\n\r\n")
		unifyNetwork(connectedComponents)
		computeAndShowGraph()
	}

}

func computeAndShowGraph() *Graph {
	graph := NewGraph()
	PrintGraph(graph.graph)
	return graph
}

// Solves network partitions.
// Network partitions are solved connecting two consecutive components in components array
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

				firstNodeConn := peerMap.connections[firstCompNode]
				secondNodeConn := peerMap.connections[secondCompNode]

				err_1 := firstNodeConn.Call("EdgePeer.AddNeighbour", secondCompNode, nil)
				err_2 := secondNodeConn.Call("EdgePeer.AddNeighbour", firstCompNode, nil)
				if err_1 != nil && err_2 != nil {
					createdAnEdge = true
				}

			}
		}
	}
}

func findNeighboursForPeer(edgePeer EdgePeer) map[EdgePeer]byte {
	// peerNum := len(graphMap.peerMap)
	createdEdge := false
	neighboursList := map[EdgePeer]byte{}
	for peer := range peerMap.heartbeats {
		// TODO Fare prima shuffle delle chiavi per non legare tutti i nodi al primo che viene restituito
		var thereIsEdge bool
		if !createdEdge {
			thereIsEdge = true
		} else {
			thereIsEdge = existsEdge()
		}
		if thereIsEdge {
			neighboursList[peer] = 0
			createdEdge = true
		}
	}
	return neighboursList
}
