package peer

import (
	"errors"
	"fmt"
	"net/rpc"
	"sync"

	bloom "github.com/tylertreat/BoomFilters"
)

type AdjacentPeers struct {
	connsMutex   sync.RWMutex
	peerConns    map[EdgePeer](*rpc.Client)
	filtersMutex sync.RWMutex
	filterMap    map[EdgePeer](*bloom.CountingBloomFilter)
}

type HeartbeatMessage struct {
	EdgePeer       EdgePeer
	NeighboursList map[EdgePeer]byte
}

type BloomFilterMessage struct {
	EdgePeer    EdgePeer
	BloomFilter *bloom.CountingBloomFilter
}

type FileRequestMessage struct {
	FileName string
	TTL      int
}

var adjacentsMap = AdjacentPeers{
	sync.RWMutex{},
	map[EdgePeer]*rpc.Client{},
	sync.RWMutex{},
	map[EdgePeer]*bloom.CountingBloomFilter{},
}

// Comunica ai tuoi vicini di aggiungerti come loro vicino
func CallAdjAddNeighbour(client *rpc.Client, neighbourPeer EdgePeer) error {

	adjacentsMap.connsMutex.Lock()
	defer adjacentsMap.connsMutex.Unlock()

	err := client.Call("EdgePeer.AddNeighbour", neighbourPeer.PeerAddr, nil)

	if err != nil {
		return errors.New("impossibile stabilire connessione con il nuovo vicino")
	}
	return nil
}

func ConnectToNode(addr string) (*rpc.Client, error) {
	client, err := rpc.DialHTTP("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("Errore Dial HTTP")
	}
	return client, nil
}
