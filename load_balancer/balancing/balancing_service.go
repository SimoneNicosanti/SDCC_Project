package balancing

import (
	"context"
	"fmt"
	"load_balancer/login"
	proto "load_balancer/proto/load_balancer"
	"math"
	"net"
	"net/http"
	"net/rpc"
	"sync"
	"time"

	"google.golang.org/grpc"
)

var balancingServer BalancingServiceServer = BalancingServiceServer{
	mapMutex:           sync.RWMutex{},
	edgeServerMap:      map[EdgeServer]int{},
	heartbeatCheckTime: time.Now(),
	heartbeats:         map[EdgeServer](time.Time){},
}

func ActAsBalancer() {
	setupRPC()
	setUpGRPC()
	PrintEvent("BALANCER_STARTED", "Waiting for connections...")
}

func (balancingServer *BalancingServiceServer) LogClient(ctx context.Context, userInfo *proto.User) (bool, error) {
	PrintEvent("CLIENT_LOGIN_ATTEMPT", "Received message, taking edge server with minimum load")
	return login.UserLogin(userInfo.Username, userInfo.Passwd), nil
}

func (balancingServer *BalancingServiceServer) GetEdge(ctx context.Context, userInfo *proto.User) (*proto.BalancerResponse, error) {
	PrintEvent("GET_EDGE_SERVER", "Received message, taking edge server with minimum load")
	success := login.UserLogin(userInfo.Username, userInfo.Passwd)
	var edgeIpAddr string
	var err error
	if success {
		edgeIpAddr, err = balancingServer.PickEdgeServer()
		if err != nil {
			return &proto.BalancerResponse{Success: false, EdgeIpAddr: ""}, err
		}
	} else {
		edgeIpAddr = ""
	}
	return &proto.BalancerResponse{Success: success, EdgeIpAddr: edgeIpAddr}, nil
}

func (balancingServer *BalancingServiceServer) PickEdgeServer() (ipAddr string, err error) {
	balancingServer.mapMutex.Lock()
	defer balancingServer.mapMutex.Unlock()
	if len(balancingServer.edgeServerMap) == 0 {
		return "", fmt.Errorf("non ci sono edge disponibili")
	}
	var minLoadEdge EdgeServer
	var minLoadValue = math.MaxInt
	for edgeServer, serverLoad := range balancingServer.edgeServerMap {
		if serverLoad < minLoadValue {
			minLoadEdge = edgeServer
		}
	}
	balancingServer.edgeServerMap[minLoadEdge]++
	return minLoadEdge.ServerAddr, nil
}

func setupRPC() {
	err := rpc.Register(&balancingServer)
	ExitOnError("Impossibile registrare il servizio", err)
	rpc.HandleHTTP()
	list, err := net.Listen("tcp", ":4321")
	ExitOnError("Impossibile mettersi in ascolto sulla porta", err)
	go http.Serve(list, nil)
	go checkHeartbeat()
}

func setUpGRPC() {
	ipAddr, err := GetMyIPAddr()
	ExitOnError("[*GRPC_SETUP_ERROR*] -> failed to retrieve balancer IP address", err)
	lis, err := net.Listen("tcp", ipAddr+":0")
	ExitOnError("[*GRPC_SETUP_ERROR*] -> failed to listen on endpoint", err)
	ExitOnError("[*GRPC_SETUP_ERROR*] -> failed to listen", err)
	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(GetIntEnvironmentVariable("MAX_GRPC_MESSAGE_SIZE")), // Imposta la nuova dimensione massima
	}
	grpcServer := grpc.NewServer(opts...)
	proto.RegisterBalancingServiceServer(grpcServer, &BalancingServiceServer{})
	PrintEvent("GRPC_SERVER_STARTED", "Grpc server started in server with endpoint : "+lis.Addr().String())
	go grpcServer.Serve(lis)
}

func (balancingServer *BalancingServiceServer) Heartbeat(heartbeatMessage HeartbeatMessage, replyPtr *int) error {
	*replyPtr = 0

	balancingServer.mapMutex.Lock()
	defer balancingServer.mapMutex.Unlock()

	edgeServer := heartbeatMessage.EdgeServer
	_, isInMap := balancingServer.heartbeats[edgeServer]
	if !isInMap {
		balancingServer.edgeServerMap[edgeServer] = heartbeatMessage.CurrentLoad
	}

	balancingServer.heartbeats[edgeServer] = time.Now()

	return nil
}

func checkHeartbeat() {
	MONITOR_TIMER := GetIntEnvironmentVariable("MONITOR_TIMER")
	HEARTBEAT_THR := GetIntEnvironmentVariable("HEARTBEAT_THR")
	for {
		time.Sleep(time.Duration(MONITOR_TIMER) * time.Second)
		checkForDeadServers(float64(HEARTBEAT_THR))
	}
}

func checkForDeadServers(heartbeatThr float64) {
	balancingServer.mapMutex.Lock()
	defer balancingServer.mapMutex.Unlock()

	lastCheckTime := balancingServer.heartbeatCheckTime
	for edgeServer, lastHeartbeatTime := range balancingServer.heartbeats {
		if lastCheckTime.Sub(lastHeartbeatTime).Seconds() > heartbeatThr {
			//Il edgeServer viene considerato caduto e viene rimosso
			PrintEvent("DEAD_EDGE_SRV_FOUND", fmt.Sprintf("Edge server '%s' è morto", edgeServer.ServerAddr))
			_, isInMap := balancingServer.edgeServerMap[edgeServer]
			if isInMap {
				delete(balancingServer.edgeServerMap, edgeServer)
			}
			delete(balancingServer.heartbeats, edgeServer)
		}
	}
	balancingServer.heartbeatCheckTime = time.Now()
}

func (balancingServer *BalancingServiceServer) SignalJobEnd(edgeServer EdgeServer, returnPtr *int) error {
	*returnPtr = 0

	balancingServer.mapMutex.Lock()
	defer balancingServer.mapMutex.Unlock()

	value, isInMap := balancingServer.edgeServerMap[edgeServer]
	if isInMap && value > 0 {
		balancingServer.edgeServerMap[edgeServer]--
	}
	return nil
}
