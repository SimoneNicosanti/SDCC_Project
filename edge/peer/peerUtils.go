package peer

import (
	"errors"
	"net/rpc"
	"sync"
)

type PeerConnection struct {
	peer   EdgePeer
	client *rpc.Client
}

type AdjacentPeers struct {
	mutex     sync.RWMutex
	peerConns []PeerConnection
}

type HeartbeatMessage struct {
	EdgePeer       EdgePeer
	NeighboursList []EdgePeer
}

var Adjacent = AdjacentPeers{sync.RWMutex{}, []PeerConnection{}}
var RegistryConn PeerConnection

// Aggiunge una connessione verso un nuovo vicino
func AddConnection(peerConn PeerConnection) {
	Adjacent.mutex.Lock()
	defer Adjacent.mutex.Unlock()

	Adjacent.peerConns = append(Adjacent.peerConns, peerConn)
}

// Comunica ai tuoi vicini di aggiungerti come loro vicino
func CallAdjAddNeighbour(client *rpc.Client, neighbourPeer EdgePeer) error {
	Adjacent.mutex.Lock()
	defer Adjacent.mutex.Unlock()

	err := client.Call("EdgePeer.AddNeighbour", neighbourPeer.PeerAddr, nil)

	if err != nil {
		return errors.New("impossibile stabilire connessione con il nuovo vicino")
	}
	return nil
}
