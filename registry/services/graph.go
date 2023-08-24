package services

func FindConnectedComponents(peerMap map[EdgePeer]([]EdgePeer)) [][]EdgePeer {
	visitedMap := make(map[EdgePeer](bool))
	for edge := range peerMap {
		visitedMap[edge] = false
	}

	// Declares slice for connected components: size is 0 because in this way len(slice) = #connectedComponents
	connectedComponents := make([][]EdgePeer, 0)
	for peer := range peerMap {
		foundedPeers := recursiveConnectedComponentsResearch(peerMap, peer, visitedMap)
		foundedPeers = append(foundedPeers, peer)
		connectedComponents = append(connectedComponents, foundedPeers)
	}

	return connectedComponents
}

// Recursive search for connected components
func recursiveConnectedComponentsResearch(peerMap map[EdgePeer]([]EdgePeer), peer EdgePeer, visitedMap map[EdgePeer](bool)) []EdgePeer {
	visitedMap[peer] = true
	foundedPeers := make([]EdgePeer, 0)
	peerNeighbours := peerMap[peer]
	for index := range peerNeighbours {
		neighbour := peerNeighbours[index]
		visited, isInMap := visitedMap[neighbour]
		// Non è detto che il nodo sia chiave nel grafo
		// Quando mi arriva heartbeat dopo ripresa del registry, il nodo mi dice i suoi vicini, ma io aggiungo solo il nodo da cui ho ricevuto heartbeat
		if !visited && isInMap {
			neighboursOfNeighbour := recursiveConnectedComponentsResearch(peerMap, neighbour, visitedMap)
			for index := range neighboursOfNeighbour {
				foundedPeers = append(foundedPeers, neighboursOfNeighbour[index])
				// TODO Check not repeated elements inside list
			}
		}
	}
	return foundedPeers
}