package peer

import (
	"edge/utils"
	"errors"
	"fmt"
	"net/rpc"
	"sync"
	"time"

	bloom "github.com/tylertreat/BoomFilters"
)

type AdjConnection struct {
	peerConnection *rpc.Client
	missedPing     int
}

type AdjacentPeers struct {
	connsMutex   sync.RWMutex
	peerConns    map[EdgePeer](AdjConnection)
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
	FileName       string
	TTL            int
	SenderPeer     EdgePeer
	TicketId       string
	ForwarderPeer  EdgePeer
	CallbackServer string
}

// Attenzione alla gestione dei Mutex
type FileRequestCache struct {
	mutex      sync.RWMutex
	messageMap map[FileRequestMessage](time.Time)
}

var adjacentsMap = AdjacentPeers{
	sync.RWMutex{},
	map[EdgePeer]AdjConnection{},
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
	call := client.Go("EdgePeer.AddNeighbour", neighbourPeer, nil, nil)
	select {
	case <-call.Done:
		if call.Error != nil {
			return errors.New("impossibile stabilire connessione con il nuovo vicino")
		}
	case <-time.After(time.Second * time.Duration(utils.GetInt64EnvironmentVariable("MAX_WAITING_TIME_FOR_EDGE"))):
		return fmt.Errorf("")
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
