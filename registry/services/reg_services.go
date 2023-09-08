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

// var connectionMap ConnectionMap = ConnectionMap{
// 	sync.RWMutex{},
// 	make(map[EdgePeer](*rpc.Client)),
// }

// var graphMap GraphMap = GraphMap{
// 	mutex:   sync.RWMutex{},
// 	peerMap: map[EdgePeer](map[EdgePeer](byte)){},
// }

// var heartbeatMap HeartbeatMap = HeartbeatMap{
// 	sync.RWMutex{},
// 	time.Now(),
// 	make(map[EdgePeer](time.Time)),
// }

var peerMap PeerMap = PeerMap{
	mutex:              sync.RWMutex{},
	connections:        map[EdgePeer]*rpc.Client{},
	heartbeatCheckTime: time.Now(),
	heartbeats:         map[EdgePeer](time.Time){},
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

	peerMap.mutex.Lock()
	defer peerMap.mutex.Unlock()

	peerConnection, err := connectToNode(edgePeer)
	if err != nil {
		return errors.New("impossibile stabilire connessione con il peer")
	}

	peerMap.connections[edgePeer] = peerConnection
	peerMap.heartbeats[edgePeer] = time.Now()

	//PrintGraph(graphMap.peerMap)
	neighboursList := findNeighboursForPeer(edgePeer)
	*replyPtr = neighboursList
	return nil
}

func (r *RegistryService) Heartbeat(heartbeatMessage HeartbeatMessage, replyPtr *int) error {
	peerMap.mutex.Lock()
	defer peerMap.mutex.Unlock()

	edgePeer := heartbeatMessage.EdgePeer
	_, ok := peerMap.heartbeats[heartbeatMessage.EdgePeer]
	if !ok {
		// Il peer non è presente nel sistema --> Era stato tolto oppure ho un recupero dal fallimento
		log.Println("Trovato Peer Attivo >>> " + edgePeer.PeerAddr + "\r\n")

		peerConn, err := connectToNode(edgePeer)
		if err != nil {
			log.Println("Errore connessione al Peer >> " + edgePeer.PeerAddr + "\r\n")
			return err
		}

		peerMap.connections[edgePeer] = peerConn
	}

	peerMap.heartbeats[edgePeer] = time.Now()

	return nil
}
