package services

import (
	"fmt"
	"log"
	"math/rand"
	"net/rpc"
	"registry/utils"
	"strconv"
	"strings"
	"sync"
	"time"
)

type RegistryService struct {
	mutex              sync.RWMutex
	connections        map[EdgePeer]*rpc.Client
	heartbeatCheckTime time.Time
	heartbeats         map[EdgePeer](time.Time)
}

type EdgePeer struct {
	PeerAddr string
}

type PeerMap struct {
	mutex              sync.RWMutex
	connections        map[EdgePeer](*rpc.Client)
	heartbeatCheckTime time.Time
	heartbeats         map[EdgePeer](time.Time)
}

type HeartbeatMessage struct {
	EdgePeer EdgePeer
}

func connectToNode(edgePeer EdgePeer) (*rpc.Client, error) {
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

func PrintGraph(peerMap map[EdgePeer]map[EdgePeer]byte) {
	if len(peerMap) == 0 {
		utils.PrintEvent("EMPTY_GRAPH", "Graph is empty. No peers in network.")
	} else {
		log.Printf("\033[1;30;47m[*CURRENT_GRAPH*]\033[0m\r\n")
		for node, neighbors := range peerMap {
			fmt.Printf("%s --> [", strings.Split(node.PeerAddr, ".")[3])

			neighborList := make([]string, 0)
			for neighbor := range neighbors {
				neighborList = append(neighborList, strings.Split(neighbor.PeerAddr, ".")[3])
			}

			fmt.Printf("%s]\n", strings.Join(neighborList, ", "))
		}
		fmt.Printf("\r\n")
	}
}
