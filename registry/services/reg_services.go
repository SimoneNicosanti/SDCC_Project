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

	fmt.Println("Waiting for connections...\r\n")

	// go monitorNetwork()
	go http.Serve(list, nil)
	go checkForDeadPeers()
	go monitorNetwork()
}

func (r *RegistryService) PeerEnter(edgePeer EdgePeer, replyPtr *map[EdgePeer]byte) error {
	log.Println("Entered " + edgePeer.PeerAddr + "\r\n")
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

	heartbeatMap.heartbeats[edgePeer] = time.Now()

	PrintGraph(graphMap.peerMap)

	*replyPtr = neighboursList
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
		*replyPtr = graphMap.peerMap[edgePeer]
	} else {
		// TODO Il peer non è presente nel sistema --> Era stato tolto oppure ho un recupero dal fallimento
		// Aggiungere l'elenco dei peer nel graphMap
		// Ritorno lo stesso di ciò che mi è stato inviato
		log.Println("Trovato Peer Attivo >>> " + edgePeer.PeerAddr + "\r\n")

		// Il nodo è ancora vivo (oppure l'ho ritrovato dopo il ripristino) quindi riapro la connessione e reimposto l'heartbeat
		client, err := connectToNode(edgePeer)
		if err != nil {
			log.Println("Errore connessione al Peer >> " + edgePeer.PeerAddr + "\r\n")
			return err
		}

		// TODO Ragionare bene su dove mettere questo ciclo
		for neighbourPeer := range heartbeatMessage.NeighboursList {
			// Reinserisco il nodo nelle connessioni dei vicini
			// Lo aggiungo solo se il nodo vicino è già presente nella mappa, altrimenti devo aspettare il suo heartbeat
			neighEdges, isInMap := graphMap.peerMap[neighbourPeer]
			if isInMap {
				neighEdges[edgePeer] = 0
			}
		}

		*replyPtr = heartbeatMessage.NeighboursList

		connectionMap.connections[edgePeer] = client
		graphMap.peerMap[edgePeer] = heartbeatMessage.NeighboursList
	}

	heartbeatMap.heartbeats[edgePeer] = time.Now()

	//log.Println("Heartbeat From >> " + edgePeer.PeerAddr + "\r\n")

	return nil
}
