package services

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"registry/utils"
	"sync"
	"time"
)

func ActAsRegistry() {

	registryService := RegistryService{
		mutex:              sync.RWMutex{},
		heartbeatCheckTime: time.Now(),
		connections:        map[EdgePeer]*rpc.Client{},
		heartbeats:         map[EdgePeer](time.Time){},
	}

	err := rpc.Register(&registryService)
	if err != nil {
		utils.ExitOnError("Impossibile registrare il servizio", err)
	}

	rpc.HandleHTTP()
	list, err := net.Listen("tcp", ":1234")
	if err != nil {
		utils.ExitOnError("Impossibile mettersi in ascolto sulla porta", err)
	}

	utils.PrintEvent("REGISTRY_STARTED", "Waiting for connections...")

	// go monitorNetwork()
	go http.Serve(list, nil)
	go (&registryService).checkHeartbeat()
	go (&registryService).monitorNetwork()
}

func (r *RegistryService) PeerEnter(edgePeer EdgePeer, replyPtr *map[EdgePeer]byte) error {
	utils.PrintEvent("PEER_ENTERED", fmt.Sprintf("Nuovo peer '%s' rilevato!", edgePeer.PeerAddr))

	r.mutex.Lock()
	defer r.mutex.Unlock()

	peerConnection, err := connectToNode(edgePeer)
	if err != nil {
		return errors.New("impossibile stabilire connessione con il peer")
	}

	r.connections[edgePeer] = peerConnection
	r.heartbeats[edgePeer] = time.Now()

	//PrintGraph(graphMap.peerMap)
	neighboursList := r.findNeighboursForPeer(edgePeer)
	*replyPtr = neighboursList
	return nil
}

func (r *RegistryService) Heartbeat(heartbeatMessage HeartbeatMessage, replyPtr *int) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	edgePeer := heartbeatMessage.EdgePeer
	_, ok := r.heartbeats[heartbeatMessage.EdgePeer]
	if !ok {
		// Il peer non è presente nel sistema --> Era stato tolto oppure ho un recupero dal fallimento
		utils.PrintEvent("ALIVE_PEER_FOUND", fmt.Sprintf("Peer '%s' è attivo!", edgePeer.PeerAddr))

		peerConn, err := connectToNode(edgePeer)
		if err != nil {
			utils.PrintEvent("PEER_CONN_ERR", fmt.Sprintf("Errore nel tentativo di connessione al peer '%s'!", edgePeer.PeerAddr))
			return err
		}

		r.connections[edgePeer] = peerConn
	}

	r.heartbeats[edgePeer] = time.Now()

	return nil
}
