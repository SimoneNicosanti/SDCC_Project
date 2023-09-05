package peer

import (
	"errors"
	"fmt"
	"net/rpc"
	"sync"
	"time"

	bloom "github.com/tylertreat/BoomFilters"
)

type AdjacentPeers struct {
	connsMutex   sync.RWMutex
	peerConns    map[EdgePeer](*rpc.Client)
	filtersMutex sync.RWMutex
	filterMap    map[EdgePeer](*bloom.StableBloomFilter)
}

type HeartbeatMessage struct {
	EdgePeer       EdgePeer
	NeighboursList map[EdgePeer]byte
}

type BloomFilterMessage struct {
	EdgePeer    EdgePeer
	BloomFilter []byte
}

type FileRequestMessage struct {
	FileName   string
	TTL        int
	SenderPeer EdgePeer
	TicketId   string
}

// Attenzione alla gestione dei Mutex
type FileRequestCache struct {
	mutex      sync.RWMutex
	messageMap map[FileRequestMessage](time.Time)
}

var adjacentsMap = AdjacentPeers{
	sync.RWMutex{},
	map[EdgePeer]*rpc.Client{},
	sync.RWMutex{},
	map[EdgePeer]*bloom.StableBloomFilter{},
}

var fileRequestCache = FileRequestCache{
	sync.RWMutex{},
	map[FileRequestMessage]time.Time{},
}

// Comunica ai tuoi vicini di aggiungerti come loro vicino
func CallAdjAddNeighbour(client *rpc.Client, neighbourPeer EdgePeer) error {

	adjacentsMap.connsMutex.Lock()
	defer adjacentsMap.connsMutex.Unlock()

	err := client.Call("EdgePeer.AddNeighbour", neighbourPeer, nil)

	if err != nil {
		return errors.New("impossibile stabilire connessione con il nuovo vicino")
	}
	return nil
}

func ConnectToNode(addr string) (*rpc.Client, error) {
	client, err := rpc.DialHTTP("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("errore Dial HTTP")
	}
	return client, nil
}
