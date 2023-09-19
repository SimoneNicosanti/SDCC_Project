package balancing

import (
	"context"
	"fmt"
	"load_balancer/login"
	"load_balancer/proto"
	"math"
	"net"
	"net/http"
	"net/rpc"
	"sync"
	"time"
)

func (server *BalancingServer) GetEdge(ctx context.Context, userInfo *proto.User) (*proto.Response, error) {
	PrintEvent("GET_EDGE", "Received message, taking edge with minimum load")
	success := login.UserLogin(userInfo.Username, userInfo.Passwd)
	var edgeIpAddr string
	var err error
	if success {
		edgeIpAddr, err = server.takeMinimumLoadEdge()
		if err != nil {
			return &proto.Response{Success: false, EdgeIpAddr: ""}, err
		}
	} else {
		edgeIpAddr = ""
	}
	return &proto.Response{Success: success, EdgeIpAddr: edgeIpAddr}, nil
}

func (server *BalancingServer) takeMinimumLoadEdge() (ipAddr string, err error) {
	server.mapMutex.Lock()
	defer server.mapMutex.Unlock()
	if len(server.peerServerMap) == 0 {
		return "", fmt.Errorf("non ci sono edge disponibili")
	}
	var minLoadEdge PeerServer
	var minLoadValue = math.MaxInt
	for edgePeer, edgeLoad := range server.peerServerMap {
		if edgeLoad < minLoadValue {
			minLoadEdge = edgePeer
		}
	}
	server.peerServerMap[minLoadEdge]++
	return minLoadEdge.PeerAddr, nil
}

func ActAsBalancer() {
	balancingServer := BalancingServer{
		mapMutex:           sync.RWMutex{},
		peerServerMap:      map[PeerServer]int{},
		heartbeatCheckTime: time.Now(),
		heartbeats:         map[PeerServer](time.Time){},
	}

	// TODO Registrare correttamente i servizi per Peer e Client
	// Uno è RPC uno è gRPC
	err := rpc.Register(&balancingServer)
	if err != nil {
		ExitOnError("Impossibile registrare il servizio", err)
	}

	rpc.HandleHTTP()
	list, err := net.Listen("tcp", ":4321")
	if err != nil {
		ExitOnError("Impossibile mettersi in ascolto sulla porta", err)
	}

	PrintEvent("BALANCER_STARTED", "Waiting for connections...")

	go http.Serve(list, nil)
	go checkHeartbeat()
}

func (server *BalancingServer) Heartbeat(heartbeatMessage HeartbeatMessage, replyPtr *int) error {
	*replyPtr = 0

	server.mapMutex.Lock()
	defer server.mapMutex.Unlock()

	peerServer := heartbeatMessage.PeerServer
	_, isInMap := server.heartbeats[peerServer]
	if !isInMap {
		server.peerServerMap[peerServer] = heartbeatMessage.CurrentLoad
	}

	server.heartbeats[peerServer] = time.Now()

	return nil
}

// func checkHeartbeat() {
// 	MONITOR_TIMER := utils.GetIntegerEnvironmentVariable("MONITOR_TIMER")
// 	HEARTBEAT_THR := utils.GetIntegerEnvironmentVariable("HEARTBEAT_THR")
// 	for {
// 		time.Sleep(time.Duration(MONITOR_TIMER) * time.Second)
// 		checkForDeadPeers(float64(HEARTBEAT_THR))
// 	}
// }

// func checkForDeadPeers(heartbeatThr float64) {
// 	peerMap.mutex.Lock()
// 	defer peerMap.mutex.Unlock()

// 	lastCheckTime := peerMap.heartbeatCheckTime
// 	for edgePeer, lastHeartbeatTime := range peerMap.heartbeats {
// 		if lastCheckTime.Sub(lastHeartbeatTime).Seconds() > heartbeatThr {
// 			//Il peer viene considerato caduto e viene rimosso dalla rete
// 			utils.PrintEvent("DEAD_PEER_FOUND", fmt.Sprintf("Peer '%s' è morto", edgePeer.PeerAddr))
// 			deadPeerConn, isInMap := peerMap.connections[edgePeer]
// 			if isInMap {
// 				deadPeerConn.Close()
// 				delete(peerMap.connections, edgePeer)
// 			}
// 			delete(peerMap.heartbeats, edgePeer)
// 			//PrintGraph(graphMap.peerMap)
// 		}
// 	}
// 	peerMap.heartbeatCheckTime = time.Now()
// }

func (server *BalancingServer) SignalJobEnd(peerServer PeerServer, returnPtr *int) error {
	server.mapMutex.Lock()
	defer server.mapMutex.Unlock()

	value, isInMap := server.peerServerMap[peerServer]
	if isInMap && value > 0 {
		server.peerServerMap[peerServer]--
	}
	return nil
}
