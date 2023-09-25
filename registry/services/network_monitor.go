package services

import (
	"fmt"
	"registry/utils"
	"strconv"
	"time"
)

func (r *RegistryService) checkHeartbeat() {
	MONITOR_TIMER := utils.GetIntEnvironmentVariable("MONITOR_TIMER")
	HEARTBEAT_THR := utils.GetIntEnvironmentVariable("HEARTBEAT_THR")
	for {
		time.Sleep(time.Duration(MONITOR_TIMER) * time.Second)
		r.checkForDeadPeers(float64(HEARTBEAT_THR))
	}
}

func (r *RegistryService) checkForDeadPeers(heartbeatThr float64) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	lastCheckTime := r.heartbeatCheckTime
	for edgePeer, lastHeartbeatTime := range r.heartbeats {
		if lastCheckTime.Sub(lastHeartbeatTime).Seconds() > heartbeatThr {
			//Il peer viene considerato caduto e viene rimosso dalla rete
			utils.PrintEvent("DEAD_PEER_FOUND", fmt.Sprintf("Peer '%s' è morto", edgePeer.PeerAddr))
			deadPeerConn, isInMap := r.connections[edgePeer]
			if isInMap {
				deadPeerConn.Close()
				delete(r.connections, edgePeer)
			}
			delete(r.heartbeats, edgePeer)
			//PrintGraph(graphMap.peerMap)
		}
	}
	r.heartbeatCheckTime = time.Now()
}

// Periodically check if network is connected
func (r *RegistryService) monitorNetwork() {
	monitorTimerString := utils.GetEnvironmentVariable("MONITOR_TIMER")
	MONITOR_TIMER, err := strconv.ParseInt(monitorTimerString, 10, 64)
	utils.ExitOnError("Impossibile fare il parsing di MONITOR_TIME", err)
	for {
		time.Sleep(time.Duration(MONITOR_TIMER) * time.Second)
		r.solveNetworkPartitions()
	}
}

// Checks for network partitions and solve them if any is found
func (r *RegistryService) solveNetworkPartitions() {
	// Is a different function because network partitions solve is necessary for PeerExit and for PeerCrash too
	r.mutex.Lock()
	defer r.mutex.Unlock()

	graph := r.NewGraph()
	PrintGraph(graph.graph)

	connectedComponents := graph.FindConnectedComponents()

	if len(connectedComponents) > 1 {
		utils.PrintEvent("PARTITION_FOUND", "Trovata partizione di rete...")
		r.unifyNetwork(connectedComponents)
	}

}

// Solves network partitions.
// Network partitions are solved connecting two consecutive components in components array
func (r *RegistryService) unifyNetwork(connectedComponents [][]EdgePeer) {

	for componentIndex := 0; componentIndex < len(connectedComponents)-1; componentIndex++ {
		firstComponent := connectedComponents[componentIndex]
		secondComponent := connectedComponents[componentIndex+1]
		r.unifyTwoComponents(firstComponent, secondComponent)
	}
}

// Unifies two components of network
func (r *RegistryService) unifyTwoComponents(firstComponent []EdgePeer, secondComponent []EdgePeer) {
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

				firstNodeConn := r.connections[firstCompNode]
				secondNodeConn := r.connections[secondCompNode]

				call_1 := firstNodeConn.Go("EdgePeer.AddNeighbour", secondCompNode, 0, nil)
				call_2 := secondNodeConn.Go("EdgePeer.AddNeighbour", firstCompNode, 0, nil)
				//controlliamo se almeno una delle due call abbia successo (l'altro verso dell'edge verrà ricostruito in un secondo momento eventualmente dopo la ricezione di un ping)
				select {
				case <-call_1.Done:
					if call_1.Error == nil {
						createdAnEdge = true
					}
				case <-call_2.Done:
					if call_2.Error == nil {
						createdAnEdge = true
					}
				case <-time.After(time.Second * time.Duration(utils.GetIntEnvironmentVariable("MAX_WAITING_TIME_FOR_REGISTRY"))):
					utils.PrintEvent("TIMEOUT_ERROR", fmt.Sprintf("Durante un tentativo di unificazione di componenti connesse non è stata ricevuta una risposta entro %d secondi da '%s' o '%s'", utils.GetIntEnvironmentVariable("MAX_WAITING_TIME_FOR_REGISTRY"), firstCompNode.PeerAddr, secondCompNode.PeerAddr))
				}
			}
		}
	}
}

func (r *RegistryService) findNeighboursForPeer(edgePeer EdgePeer) map[EdgePeer]byte {
	createdEdge := false
	neighboursList := map[EdgePeer]byte{}
	for peer := range r.heartbeats {
		if peer == edgePeer {
			continue
		}
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
