package peer

import (
	"errors"
	"net/rpc"
	"sync"
)

type AdjacentPeers struct {
	mutex     sync.RWMutex
	peerConns map[EdgePeer](*rpc.Client)
}

type HeartbeatMessage struct {
	EdgePeer       EdgePeer
	NeighboursList map[EdgePeer]byte
}

var adjacentsMap = AdjacentPeers{sync.RWMutex{}, map[EdgePeer]*rpc.Client{}}

// Aggiunge una connessione verso un nuovo vicino
func addConnection(edgePeer EdgePeer, conn *rpc.Client) {
	adjacentsMap.mutex.Lock()
	defer adjacentsMap.mutex.Unlock()
	adjacentsMap.peerConns[edgePeer] = conn
}

// Comunica ai tuoi vicini di aggiungerti come loro vicino
func CallAdjAddNeighbour(client *rpc.Client, neighbourPeer EdgePeer) error {
	adjacentsMap.mutex.Lock()
	defer adjacentsMap.mutex.Unlock()

	err := client.Call("EdgePeer.AddNeighbour", neighbourPeer.PeerAddr, nil)

	if err != nil {
		return errors.New("impossibile stabilire connessione con il nuovo vicino")
	}
	return nil
}

func connectToNode(addr string) (*rpc.Client, string, error) {
	client, err := rpc.DialHTTP("tcp", addr)
	if err != nil {
		return nil, "Errore Dial HTTP", err
	}
	return client, "", err
}
