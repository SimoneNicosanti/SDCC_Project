package services

import (
	"math/rand"
	"net/rpc"
	"registry/utils"
	"strconv"
	"sync"
	"time"
)

type RegistryService int

type GraphMap struct {
	mutex   sync.RWMutex
	peerMap map[EdgePeer]([]EdgePeer)
}

type EdgePeer struct {
	PeerAddr string
}

type ConnectionMap struct {
	mutex       sync.RWMutex
	connections map[EdgePeer](*rpc.Client)
}

type HeartbeatMap struct {
	mutex       sync.RWMutex
	lastChecked time.Time
	heartbeats  map[EdgePeer](time.Time)
}

type HeartbeatMessage struct {
	EdgePeer       EdgePeer
	NeighboursList []EdgePeer
}

func connectToPeer(edgePeer EdgePeer) (*rpc.Client, error) {
	client, err := rpc.DialHTTP("tcp", edgePeer.PeerAddr)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func existsEdge() bool {
	randomNumber := rand.Float64()

	RAND_THR, err := strconv.ParseFloat(utils.GetEnvironmentVariable("RAND_THR"), 64)
	utils.ExitOnError("Impossibile fare il parsing di RAND_THR", err)
	return randomNumber > RAND_THR
}
