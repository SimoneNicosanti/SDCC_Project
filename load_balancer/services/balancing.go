package services

import (
	"fmt"
	"load_balancer/engineering"
	proto "load_balancer/proto/load_balancer"
	"load_balancer/utils"
	"net"
	"net/http"
	"net/rpc"
	"time"

	"google.golang.org/grpc"
)

func ActAsBalancer() {
	setupRPC()
	setUpGRPC()
	engineering.PutOnRedis("edoardo", "andrea")
	utils.PrintEvent("BALANCER_STARTED", "Waiting for connections...")
}

func setupRPC() {
	err := rpc.Register(&balancingServer)
	utils.ExitOnError("Impossibile registrare il servizio", err)
	rpc.HandleHTTP()
	list, err := net.Listen("tcp", ":4321")
	utils.ExitOnError("Impossibile mettersi in ascolto sulla porta", err)
	go http.Serve(list, nil)
	go checkHeartbeat()
}

func setUpGRPC() {
	ipAddr, err := utils.GetMyIPAddr()
	utils.ExitOnError("[*GRPC_SETUP_ERROR*] -> failed to retrieve balancer IP address", err)
	lis, err := net.Listen("tcp", ipAddr+":5432")
	utils.ExitOnError("[*GRPC_SETUP_ERROR*] -> failed to listen on endpoint", err)
	utils.ExitOnError("[*GRPC_SETUP_ERROR*] -> failed to listen", err)
	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(utils.GetIntEnvironmentVariable("MAX_GRPC_MESSAGE_SIZE")), // Imposta la nuova dimensione massima
	}
	grpcServer := grpc.NewServer(opts...)
	proto.RegisterBalancingServiceServer(grpcServer, &BalancingServiceServer{})
	utils.PrintEvent("GRPC_SERVER_STARTED", "Grpc server started in server with endpoint : "+lis.Addr().String())
	go grpcServer.Serve(lis)
}

func checkHeartbeat() {
	MONITOR_TIMER := utils.GetIntEnvironmentVariable("MONITOR_TIMER")
	HEARTBEAT_THR := utils.GetIntEnvironmentVariable("HEARTBEAT_THR")
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
			utils.PrintEvent("DEAD_EDGE_SRV_FOUND", fmt.Sprintf("Edge server '%s' è morto", edgeServer.ServerAddr))
			_, isInMap := balancingServer.edgeServerMap[edgeServer]
			if isInMap {
				delete(balancingServer.edgeServerMap, edgeServer)
			}
			delete(balancingServer.heartbeats, edgeServer)
		}
	}
	balancingServer.heartbeatCheckTime = time.Now()
}
