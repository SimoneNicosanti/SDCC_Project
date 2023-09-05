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
				log.Println("Trovato Peer Morto >>> " + edgePeer.PeerAddr + "\r\n")
				removeDeadNode(edgePeer)
				PrintGraph(graphMap.peerMap)
			}
		}
		heartbeatMap.lastChecked = time.Now()

		// Se c'è stato il recupero, allora tutti quanti devono avermi mandato heartbeat da quando ho recuperato
		// Se qualcuno non l'ha fatto ed è caduto, potrebbe essere presente nella lista degli archi mandata dagli altri
		// Quindi devo rimuoverlo dal grafo
		// Nelle altre strutture dati non c'è perché viene aperta connessione / aperto heartbeat solo se è stato
		// ricevuti il primo heartbeat
		// TODO Ricontrolla se serve (NB: potrebbe servire per evitare inconsistenza ed eventuali controlli errati sulle componenti connesse)
		for _, peerEdges := range graphMap.peerMap {
			for neighbourPeer := range peerEdges {
				_, isInMap := graphMap.peerMap[neighbourPeer]
				if !isInMap {
					removeDeadNode(neighbourPeer)
				}
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

	for _, otherPeerConnections := range graphMap.peerMap {
		delete(otherPeerConnections, deadPeer)
	}

	deadPeerConn, _ := connectionMap.connections[deadPeer]
	deadPeerConn.Close()
	delete(connectionMap.connections, deadPeer)

	delete(heartbeatMap.heartbeats, deadPeer)
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
		log.Printf("Trovata partizione di rete\r\n\r\n")
		unifyNetwork(connectedComponents)
		PrintGraph(graphMap.peerMap)
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
				graphMap.peerMap[firstCompNode][secondCompNode] = 0
				graphMap.peerMap[secondCompNode][firstCompNode] = 0

				firstNodeConn := connectionMap.connections[firstCompNode]
				secondNodeConn := connectionMap.connections[secondCompNode]

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
	for peer := range graphMap.peerMap {
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
