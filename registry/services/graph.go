package services

import (
	"fmt"
	"registry/utils"
	"time"
)

type Graph struct {
	graph map[EdgePeer](map[EdgePeer]byte)
}

func NewGraph() *Graph {
	graph := new(Graph)
	graph.graph = map[EdgePeer](map[EdgePeer]byte){}
	buildGraph(graph)
	return graph
}

func buildGraph(graph *Graph) {
	// Ottenimento dello stato attuale della rete
	for peer, conn := range peerMap.connections {
		_, hasHeartbeat := peerMap.heartbeats[peer]
		if hasHeartbeat {
			neighboursPtr := new(map[EdgePeer]byte)
			call := conn.Go("EdgePeer.GetNeighbours", 0, neighboursPtr, nil)
			select {
			case <-call.Done:
				if call.Error != nil {
					utils.PrintEvent("GRAPH_ERROR", fmt.Sprintf("Ottenimento vicini del nodo '%s' non riuscita", peer.PeerAddr))
				}
				graph.AddNodeAndConnections(peer, *neighboursPtr)
			case <-time.After(time.Second * time.Duration(utils.GetIntEnvironmentVariable("MAX_WAITING_TIME_FOR_REGISTRY"))):
				utils.PrintEvent("TIMEOUT_ERROR", fmt.Sprintf("Non è stata ricevuta una risposta entro %d secondi da '%s'", utils.GetIntEnvironmentVariable("MAX_WAITING_TIME_FOR_REGISTRY"), peer.PeerAddr))
			}
		}
	}

	// Rimozione dei nodi che non hanno heartbeat
	for _, nodeAdjs := range graph.graph {
		for adj := range nodeAdjs {
			_, hasHeartbeat := peerMap.heartbeats[adj]
			if !hasHeartbeat {
				delete(nodeAdjs, adj)
			}
		}
	}

	// Trasformiamo il grafo diretto in uno non diretto
	for node, nodeAdjs := range graph.graph {
		for adj := range nodeAdjs {
			_, isInGraph := graph.graph[adj]
			if isInGraph {
				graph.graph[adj][node] = 0
			}
		}
	}
}

func (g *Graph) AddNodeAndConnections(node EdgePeer, neighbours map[EdgePeer]byte) {
	g.graph[node] = neighbours
}

func (g *Graph) FindConnectedComponents() [][]EdgePeer {
	visitedMap := make(map[EdgePeer](bool))
	for edge := range g.graph {
		visitedMap[edge] = false
	}

	// Declares slice for connected components: size is 0 because in this way len(slice) = #connectedComponents
	connectedComponents := make([][]EdgePeer, 0)
	for peer, visited := range visitedMap {
		if !visited {
			foundPeers := g.recursiveConnectedComponentsResearch(peer, visitedMap)
			foundPeers = append(foundPeers, peer)
			connectedComponents = append(connectedComponents, foundPeers)
		}
	}

	return connectedComponents
}

// Recursive search for connected components
func (g *Graph) recursiveConnectedComponentsResearch(peer EdgePeer, visitedMap map[EdgePeer](bool)) []EdgePeer {
	visitedMap[peer] = true
	foundPeers := make([]EdgePeer, 0)
	peerNeighbours := g.graph[peer]
	for neighbour := range peerNeighbours {
		visited, isInMap := visitedMap[neighbour]
		// Non è detto che il nodo sia chiave nel grafo
		// Quando mi arriva heartbeat dopo ripresa del registry, il nodo mi dice i suoi vicini, ma io aggiungo solo il nodo da cui ho ricevuto heartbeat
		if !visited && isInMap {
			neighboursOfNeighbour := g.recursiveConnectedComponentsResearch(neighbour, visitedMap)
			for index := range neighboursOfNeighbour {
				foundPeers = append(foundPeers, neighboursOfNeighbour[index])
				// TODO Controlla che gli elementi nella lista non siano duplicati
			}
		}
	}
	return foundPeers
}
