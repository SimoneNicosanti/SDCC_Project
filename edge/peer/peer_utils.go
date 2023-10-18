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
	connsMutex   sync.RWMutex                            // Mutex per accedere alla mappa delle connessioni con i vicini
	peerConns    map[EdgePeer](AdjConnection)            // Mappa delle connessioni con gli edge vicini
	filtersMutex sync.RWMutex                            // Filtri di Bloom e connessioni dei vicini non sono usati sempre insieme, quindi sono presenti due mutex differenti
	filterMap    map[EdgePeer](*bloom.StableBloomFilter) // Mappa dei filtri di Bloom dei vicini
}

type BloomFilterMessage struct {
	EdgePeer    EdgePeer
	BloomFilter []byte
}

type HeartbeatMessage struct {
	EdgePeer       EdgePeer
	NeighboursList map[EdgePeer]byte
}

var adjacentsMap = AdjacentPeers{
	sync.RWMutex{},
	map[EdgePeer]AdjConnection{},
	sync.RWMutex{},
	map[EdgePeer]*bloom.StableBloomFilter{},
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
