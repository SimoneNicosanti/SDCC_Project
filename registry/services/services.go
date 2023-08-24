package services

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"registry/utils"
	"sync"
	"time"
)

var connectionMap ConnectionMap = ConnectionMap{sync.RWMutex{}, make(map[EdgePeer](*rpc.Client))}

var graphMap GraphMap = GraphMap{sync.RWMutex{}, make(map[EdgePeer]([]EdgePeer))}

var heartbeatMap HeartbeatMap = HeartbeatMap{sync.RWMutex{}, time.Now(), make(map[EdgePeer]float64)])}

func ActAsRegistry() {
	registryService := new(RegistryService)
	err := rpc.Register(registryService)
	if err != nil {
		utils.ExitOnError("Impossibile registrare il servizio", err)
	}

	rpc.HandleHTTP()
	list, err := net.Listen("tcp", ":1234")
	if err != nil {
		utils.ExitOnError("Impossibile mettersi in ascolto sulla porta", err)
	}

	fmt.Println("Waiting for connections...")

	// go monitorNetwork()
	go http.Serve(list, nil)
	go checkForDeadPeers()
}

func (r *RegistryService) PeerEnter(edgePeer EdgePeer, replyPtr *[]EdgePeer) error {
	log.Println("Entered " + edgePeer.PeerAddr)
	graphMap.mutex.Lock()
	defer graphMap.mutex.Unlock()

	peerConnection, err := connectToPeer(edgePeer)
	if err != nil {
		return errors.New("impossibile stabilire connessione con il peer")
	}
	connectionMap.mutex.Lock()
	connectionMap.connections[edgePeer] = peerConnection
	connectionMap.mutex.Unlock()

	neighboursList := findNeighboursForPeer(edgePeer)
	graphMap.peerMap[edgePeer] = neighboursList

	for index := range neighboursList {
		neighbour := neighboursList[index]
		graphMap.peerMap[neighbour] = append(graphMap.peerMap[neighbour], edgePeer)
	}
	log.Println(graphMap.peerMap)

	*replyPtr = neighboursList
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

func (r *RegistryService) PeerHeartbeat(edgePeer EdgePeer, replyPtr *int) error {
	return nil
}

func (r *RegistryService) PeerPing(inputInt int, replyPtr *int) error {
	return nil
}

func (r *RegistryService) Heartbeat(heartbeatMessage HeartbeatMessage, replyPtr *[]EdgePeer) error {
	heartbeatMap.mutex.Lock()
	graphMap.mutex.Lock()
	connectionMap.mutex.Lock()
	defer connectionMap.mutex.Unlock()
	defer graphMap.mutex.Unlock()
	defer heartbeatMap.mutex.Unlock()

	edgePeer := heartbeatMessage.EdgePeer
	_, ok := heartbeatMap[edgePeer]
	if (ok) {
		// Il peer è presente nel sistema --> Ritorno la lista dei suoi peer come la conosce il registry
		// Notare che le liste sono coerenti perché l'unico caso in cui un peer è rimosso è quando l'heartbeat non è ricevuto per un certo tempo
		*retryPtr = graphMap.peerMap[edgePeer]
	} else {
		// TODO Il peer non è presente nel sistema --> Era stato tolto oppure ho un recupero dal fallimento
		// Aggiungere l'elenco dei peer nel graphMap
		// Ritorno lo stesso di ciò che mi è stato inviato
		// Devo stabilire la connessione verso questo peer

		graphMap[edgePeer] = heartbeatMessage.NeighboursList
		client, err := connectToPeer(edgePeer)
		if (err != nil) {
			log.Println("Errore connessione al Peer >> " + edgePeer.PeerAddr)
		}
		connectionMap[edgePeer] = client

		*retryPtr = heartbeatMessage.NeighboursList
	}
	now = Time.Now()
	heartbeatMap[edgePeer] = now
}
