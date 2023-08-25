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

var connectionMap ConnectionMap = ConnectionMap{
	sync.RWMutex{},
	make(map[EdgePeer](*rpc.Client)),
}

var graphMap GraphMap = GraphMap{
	mutex:   sync.RWMutex{},
	peerMap: map[EdgePeer](map[EdgePeer](byte)){},
}

var heartbeatMap HeartbeatMap = HeartbeatMap{
	sync.RWMutex{},
	time.Now(),
	make(map[EdgePeer](time.Time)),
}

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
	go monitorNetwork()
}

func (r *RegistryService) PeerEnter(edgePeer EdgePeer, replyPtr *map[EdgePeer]byte) error {
	log.Println("Entered " + edgePeer.PeerAddr)
	graphMap.mutex.Lock()
	connectionMap.mutex.Lock()
	heartbeatMap.mutex.Lock()

	defer heartbeatMap.mutex.Unlock()
	defer connectionMap.mutex.Unlock()
	defer graphMap.mutex.Unlock()

	peerConnection, err := connectToNode(edgePeer)
	if err != nil {
		return errors.New("impossibile stabilire connessione con il peer")
	}

	connectionMap.connections[edgePeer] = peerConnection

	neighboursList := findNeighboursForPeer(edgePeer)
	graphMap.peerMap[edgePeer] = neighboursList

	for neighbour := range neighboursList {
		graphMap.peerMap[neighbour][edgePeer] = 0
	}
	log.Println(graphMap.peerMap)

	heartbeatMap.heartbeats[edgePeer] = time.Now()

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

func (r *RegistryService) Heartbeat(heartbeatMessage HeartbeatMessage, replyPtr *map[EdgePeer]byte) error {
	heartbeatMap.mutex.Lock()
	graphMap.mutex.Lock()
	connectionMap.mutex.Lock()
	defer connectionMap.mutex.Unlock()
	defer graphMap.mutex.Unlock()
	defer heartbeatMap.mutex.Unlock()

	edgePeer := heartbeatMessage.EdgePeer
	_, ok := heartbeatMap.heartbeats[edgePeer]
	if ok {
		// Il peer è presente nel sistema --> Ritorno la lista dei suoi peer come la conosce il registry
		// Notare che le liste sono coerenti perché l'unico caso in cui un peer è rimosso è quando l'heartbeat non è ricevuto per un certo tempo
		*replyPtr = graphMap.peerMap[edgePeer]
	} else {
		// TODO Il peer non è presente nel sistema --> Era stato tolto oppure ho un recupero dal fallimento
		// Aggiungere l'elenco dei peer nel graphMap
		// Ritorno lo stesso di ciò che mi è stato inviato
		log.Println("Trovato Peer Attivo >>> " + edgePeer.PeerAddr)
		graphMap.peerMap[edgePeer] = heartbeatMessage.NeighboursList

		// Lo aggiungo solo se quel nodo è già presente nella mappa, altrimenti devo aspettare il suo heartbeat
		for neighbourPeer := range heartbeatMessage.NeighboursList {
			// Reinserisco il nodo nelle connessioni dei vicini
			neighEdges, isInMap := graphMap.peerMap[neighbourPeer]
			if isInMap {
				neighEdges[edgePeer] = 0
			}
		}

		// Il nodo è ancora vivo (oppure l'ho ritrovato dopo il ripristino) quindi riapro la connessione e reimposto l'heartbeat
		client, err := connectToNode(edgePeer)
		if err != nil {
			// TODO Stabilire cosa fare in caso di errore
			log.Println("Errore connessione al Peer >> " + edgePeer.PeerAddr)
		} else {
			connectionMap.connections[edgePeer] = client
		}

		*replyPtr = heartbeatMessage.NeighboursList
	}
	heartbeatMap.heartbeats[edgePeer] = time.Now()

	log.Println("Heartbeat From >> " + edgePeer.PeerAddr)

	return nil
}
