package main

import "errors"

func (r *RegistryService) PeerEnter(edgePeer EdgePeer, replyPtr *[]EdgePeer) error {

	graphMap.mutex.Lock()
	defer graphMap.mutex.Unlock()

	peerConnection, err := connectToPeer(edgePeer)
	if err != nil {
		return errors.New("impossibile stabilire connessione con il peer")
	}
	connectionMap.mutex.Lock()
	connectionMap.connections[edgePeer] = peerConnection
	graphMap.mutex.Unlock()

	neighboursList := findNeighboursForPeer(edgePeer)
	graphMap.peerMap[edgePeer] = neighboursList

	for index := range neighboursList {
		neighbour := neighboursList[index]
		graphMap.peerMap[neighbour] = append(graphMap.peerMap[neighbour], edgePeer)
	}

	replyPtr = &neighboursList
	return nil
}

func (r *RegistryService) PeerExit(edgePeer EdgePeer, replyPtr *int) error {
	/*
		1. Rimuovere dalla lista di connessioni chiudendo la connessione
		2. Rimuovere dalla struttura del grafo rimuovendo anche da tutte le mappe in cui compare
		3. Comunicare uscita a tutti i peer che lo avevano in connessione, oppure è il nodo stesso che lo comunica ai suoi vicini ??)
		4. Valutare partizioni di rete
	*/
	return nil
}

func (r *RegistryService) PeerPing(inputInt int, replyPtr *int) error {
	return nil
}
