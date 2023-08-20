package main

// TODO Generalize functions: add params
func findConnectedComponents() [][]EdgePeer {

	visitedMap := make(map[EdgePeer](bool))
	for edge := range graphMap.peerMap {
		visitedMap[edge] = false
	}

	connectedComponents := make([][]EdgePeer, 0)
	for peer := range graphMap.peerMap {
		foundedPeers := recursiveConnectedComponentsResearch(peer, visitedMap)
		foundedPeers = append(foundedPeers, peer)
		connectedComponents = append(connectedComponents, foundedPeers)
	}

	return connectedComponents
}

func recursiveConnectedComponentsResearch(peer EdgePeer, visitedMap map[EdgePeer](bool)) []EdgePeer {
	visitedMap[peer] = true
	foundedPeers := make([]EdgePeer, 0)
	peerNeighbours := graphMap.peerMap[peer]
	for index := range peerNeighbours {
		neighbour := peerNeighbours[index]
		if !visitedMap[neighbour] {
			neighboursOfNeighbour := recursiveConnectedComponentsResearch(neighbour, visitedMap)
			for index := range neighboursOfNeighbour {
				foundedPeers = append(foundedPeers, neighboursOfNeighbour[index])
			}
		}
	}
	return foundedPeers
}
