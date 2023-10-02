package services

import (
	"fmt"
	proto "load_balancer/proto/load_balancer"
	"load_balancer/redisDB"
	"load_balancer/utils"
	"math"
	"net"
	"net/http"
	"net/rpc"
	"time"

	"google.golang.org/grpc"
)

func ActAsBalancer() {
	balancingServer.sequenceNumber = utils.GenerateUniqueRandomID()
	setupRPC()
	setUpGRPC()
	redisDB.PutOnRedis("edoardo", "edoardo")
	redisDB.PutOnRedis("andrea", "andrea")
	redisDB.PutOnRedis("simone", "simone")
	utils.PrintEvent("BALANCER_STARTED", "Aspettando connessioni...")
	go checkHeartbeat()
}

func setupRPC() {
	err := rpc.Register(&balancingServer)
	utils.ExitOnError("Impossibile registrare il servizio", err)
	rpc.HandleHTTP()
	list, err := net.Listen("tcp", ":4321")
	utils.ExitOnError("Impossibile mettersi in ascolto sulla porta", err)
	go http.Serve(list, nil)
}

func setUpGRPC() {
	ipAddr, err := utils.GetMyIPAddr()
	utils.ExitOnError("[*GRPC_SETUP_ERROR*] -> Impossibile ottenere l'indirizzo IP del load balancer", err)
	lis, err := net.Listen("tcp", ipAddr+":5432")
	utils.ExitOnError(fmt.Sprintf("[*GRPC_SETUP_ERROR*] -> Impossibile mettersi in ascolto sull'indirizzo '%s'", lis.Addr().String()), err)
	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(utils.GetIntEnvironmentVariable("MAX_GRPC_MESSAGE_SIZE")), // Imposta la nuova dimensione massima
	}
	grpcServer := grpc.NewServer(opts...)
	proto.RegisterBalancingServiceServer(grpcServer, &balancingServer)
	utils.PrintEvent("GRPC_SERVER_STARTED", "Il server GRPC è stato avviato all'indirizzo : "+lis.Addr().String())
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
			//L'edgeServer viene considerato caduto e viene rimosso
			utils.PrintEvent("DEAD_EDGE_SRV_FOUND", fmt.Sprintf("Edge server '%s' è morto", edgeServer.ServerAddr))
			_, isInMap := balancingServer.edgeServerMap[edgeServer]
			if isInMap {
				delete(balancingServer.edgeServerMap, edgeServer)
			}
			delete(balancingServer.heartbeats, edgeServer)
			convertAndPrintEdgeServerMap(balancingServer.edgeServerMap)
		}
	}
	balancingServer.heartbeatCheckTime = time.Now()
}

func (balancingServer *BalancingServiceServer) pickEdgeServer() (ipAddr string, err error) {
	balancingServer.mapMutex.Lock()
	defer balancingServer.mapMutex.Unlock()
	if len(balancingServer.edgeServerMap) == 0 {
		return "", fmt.Errorf("non ci sono edge servers disponibili")
	}

	var minLoadEdgeSlice []EdgeServer
	var minLoadEdge EdgeServer
	var minLoadValue = math.MaxInt
	for edgeServer, serverLoad := range balancingServer.edgeServerMap {
		if serverLoad < minLoadValue {
			minLoadEdge = edgeServer
			minLoadValue = serverLoad
			minLoadEdgeSlice = []EdgeServer{edgeServer}
		} else if serverLoad == minLoadValue {
			minLoadEdgeSlice = append(minLoadEdgeSlice, edgeServer)
		}
	}
	// Selezione randomica di uno degli edge con carico minimo
	balancingServer.sequenceMutex.RLock()
	minIndex := balancingServer.sequenceNumber % int64(len(minLoadEdgeSlice))
	balancingServer.sequenceMutex.RUnlock()

	minLoadEdge = minLoadEdgeSlice[minIndex]
	balancingServer.edgeServerMap[minLoadEdge] = balancingServer.edgeServerMap[minLoadEdge] + 1
	return minLoadEdge.ServerAddr, nil
}

func convertAndPrintEdgeServerMap(edgeServerMap map[EdgeServer]int) {
	printableMap := map[string]int{}
	for edgeServer, currentLoad := range edgeServerMap {
		printableMap[edgeServer.ServerAddr] = currentLoad
	}
	utils.PrintCustomMap(printableMap, "No edge servers...", "Edge Servers e rispettivo carico attuale", "CURRENT_EDGE_SERVERS", true)
}
