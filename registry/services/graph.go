package services

// TODO Controllare partizioni di rete
func FindConnectedComponents(peerMap map[EdgePeer](map[EdgePeer]byte)) [][]EdgePeer {
	visitedMap := make(map[EdgePeer](bool))
	for edge := range peerMap {
		visitedMap[edge] = false
	}

	// Declares slice for connected components: size is 0 because in this way len(slice) = #connectedComponents
	connectedComponents := make([][]EdgePeer, 0)
	for peer, visited := range visitedMap {
		if !visited {
			foundPeers := recursiveConnectedComponentsResearch(peerMap, peer, visitedMap)
			foundPeers = append(foundPeers, peer)
			connectedComponents = append(connectedComponents, foundPeers)
		}
	}

	return connectedComponents
}

// Recursive search for connected components
func recursiveConnectedComponentsResearch(peerMap map[EdgePeer](map[EdgePeer]byte), peer EdgePeer, visitedMap map[EdgePeer](bool)) []EdgePeer {
	visitedMap[peer] = true
	foundPeers := make([]EdgePeer, 0)
	peerNeighbours := peerMap[peer]
	for neighbour := range peerNeighbours {
		visited, isInMap := visitedMap[neighbour]
		// Non è detto che il nodo sia chiave nel grafo
		// Quando mi arriva heartbeat dopo ripresa del registry, il nodo mi dice i suoi vicini, ma io aggiungo solo il nodo da cui ho ricevuto heartbeat
		if !visited && isInMap {
			neighboursOfNeighbour := recursiveConnectedComponentsResearch(peerMap, neighbour, visitedMap)
			for index := range neighboursOfNeighbour {
				foundPeers = append(foundPeers, neighboursOfNeighbour[index])
				// TODO Check not repeated elements inside list
			}
		}
	}
	return foundPeers
}
