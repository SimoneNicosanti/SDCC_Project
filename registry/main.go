package main

import (
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
	"sync"
	"time"
)

var connectionMap ConnectionMap = ConnectionMap{sync.RWMutex{}, make(map[EdgePeer](*rpc.Client))}

var graphMap GraphMap = GraphMap{sync.RWMutex{}, make(map[EdgePeer]([]EdgePeer))}

func main() {

	SetupEnvVariables("conf.properties")

	registryService := new(RegistryService)
	err := rpc.Register(registryService)
	if err != nil {
		ExitOnError("Impossibile registrare il servizio", err)
	}

	rpc.HandleHTTP()
	list, err := net.Listen("tcp", ":1234")
	if err != nil {
		ExitOnError("Impossibile mettersi in ascolto sulla porta", err)
	}

	fmt.Println("Waiting for connections...")

	go monitorNetwork()
	for {
		http.Serve(list, nil)
	}
}

// Periodically check if network is connected
func monitorNetwork() {
	monitorTimerString := GetEnvironmentVariable("MONITOR_TIMER")
	MONITOR_TIMER, err := strconv.ParseInt(monitorTimerString, 10, 64)
	ExitOnError("Impossibile fare il parsing di MONITOR_TIME", err)
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
	neighboursList := []EdgePeer{}
	for peer := range graphMap.peerMap {
		if existsEdge() {
			neighboursList = append(neighboursList, peer)
		}
	}
	return neighboursList
}
