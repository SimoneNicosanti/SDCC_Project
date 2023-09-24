package server

import (
	"edge/cache"
	"edge/proto/file_transfer"
	"edge/utils"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"time"

	"google.golang.org/grpc"
)

var edgeServer EdgeServer
var workload int = 0
var balancerConnection *rpc.Client = nil

func ActAsServer() {
	cache.GetCache().StartCache()
	setUpGRPCForClients()
	go heartbeatToBalancer()
	var forever chan struct{}
	<-forever
}

func heartbeatToBalancer() {
	utils.PrintEvent("HEARTBEAT_STARTED", "Inizio meccanismo di heartbeat verso il LoadBalancer")
	HEARTBEAT_FREQUENCY := utils.GetIntEnvironmentVariable("HEARTBEAT_FREQUENCY")
	for {
		heartbeatFunction()
		time.Sleep(time.Duration(HEARTBEAT_FREQUENCY) * time.Second)
	}
}

func heartbeatFunction() {
	heartbeatMessage := HeartbeatMessage{EdgeServer: edgeServer, CurrentLoad: workload}
	if balancerConnection == nil {
		balancerConnection = utils.EnsureConnectionToTarget("load_balancer", utils.GetEnvironmentVariable("LOAD_BALANCER_ADDR"), utils.GetIntEnvironmentVariable("SLEEP_TIME_TO_RECONNECT"))
	}
	call := balancerConnection.Go("BalancingServiceServer.Heartbeat", heartbeatMessage, new(int), nil)
	select {
	case <-call.Done:
		if call.Error != nil {
			utils.PrintEvent("HEARTBEAT_ERROR", "Invio di heartbeat al Load Balancer fallito")
			log.Println(call.Error.Error())
			balancerConnection.Close()
			balancerConnection = nil
		}
	case <-time.After(time.Second * time.Duration(utils.GetInt64EnvironmentVariable("MAX_WAITING_TIME_FOR_SERVER"))):
		utils.PrintEvent("HEARTBEAT_ERROR", "Timer scaduto.. Impossibile contattare il Load Balancer")
		log.Println(call.Error.Error())
		balancerConnection.Close()
		balancerConnection = nil
	}

}

func notifyJobEnd() {
	if balancerConnection == nil {
		newBalancerConnection, err := utils.ConnectToNode(utils.GetEnvironmentVariable("LOAD_BALANCER_ADDR"))
		if err != nil {
			utils.PrintEvent("LOADBALANCER_UNREACHABLE", "Impossibile stabilire connessione con il Load Balancer")
			return
		}
		balancerConnection = newBalancerConnection
	}
	call := balancerConnection.Go("BalancingServiceServer.NotifyJobEnd", edgeServer, new(int), nil)
	select {
	case <-call.Done:
		if call.Error != nil {
			utils.PrintEvent("JOB_END_ERROR", "Notifica di job end al Load Balancer fallita")
			log.Println(call.Error.Error())
			balancerConnection.Close()
			balancerConnection = nil
		}
	case <-time.After(time.Second * time.Duration(utils.GetInt64EnvironmentVariable("MAX_WAITING_TIME_FOR_SERVER"))):
		utils.PrintEvent("JOB_END_ERROR", "Timer scaduto.. Impossibile contattare il Load Balancer")
		log.Println(call.Error.Error())
		balancerConnection.Close()
		balancerConnection = nil
	}
}

func setUpGRPCForClients() {
	ipAddr, err := utils.GetMyIPAddr()
	utils.ExitOnError("[*GRPC_SETUP_ERROR*] -> impossibile ottenere l'indirizzo ip per l'edge server", err)
	lis, err := net.Listen("tcp", ipAddr+":0")
	//Otteniamo l'indirizzo usato
	edgeServer = EdgeServer{lis.Addr().String()}
	utils.ExitOnError(fmt.Sprintf("[*GRPC_SETUP_ERROR*] -> impossibile mettersi in ascolto sull'indirizzo '%s'", lis.Addr().String()), err)
	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(utils.GetIntEnvironmentVariable("MAX_GRPC_MESSAGE_SIZE")), // Imposta la nuova dimensione massima
	}
	grpcServer := grpc.NewServer(opts...)
	file_transfer.RegisterFileServiceServer(grpcServer, &FileServiceServer{})
	utils.PrintEvent("GRPC_EDGESERVER_STARTED", "Il server GRPC per file transfer è iniziato : "+edgeServer.ServerAddr)
	go grpcServer.Serve(lis)
}
