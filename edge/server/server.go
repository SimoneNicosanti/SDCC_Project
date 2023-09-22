package server

import (
	"edge/cache"
	"edge/proto/file_transfer"
	"edge/utils"
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
	setUpGRPCForFileTransfer()
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
		newBalancerConnection, err := utils.ConnectToNode("load_balancer:4321")
		if err != nil {
			utils.PrintEvent("LOADBALANCER_UNREACHABLE", "Impossibile stabilire connessione con il Load Balancer")
			return
		}
		balancerConnection = newBalancerConnection
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
		newBalancerConnection, err := utils.ConnectToNode("load_balancer:4321")
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

func setUpGRPCForFileTransfer() {
	ipAddr, err := utils.GetMyIPAddr()
	utils.ExitOnError("[*GRPC_SETUP_ERROR*] -> failed to retrieve server IP address", err)
	lis, err := net.Listen("tcp", ipAddr+":0")
	utils.ExitOnError("[*GRPC_SETUP_ERROR*] -> failed to listen on endpoint", err)
	//Otteniamo l'indirizzo usato
	edgeServer = EdgeServer{lis.Addr().String()}
	utils.ExitOnError("[*GRPC_SETUP_ERROR*] -> failed to listen", err)
	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(utils.GetIntEnvironmentVariable("MAX_GRPC_MESSAGE_SIZE")), // Imposta la nuova dimensione massima
	}
	grpcServer := grpc.NewServer(opts...)
	file_transfer.RegisterFileServiceServer(grpcServer, &FileServiceServer{})
	utils.PrintEvent("GRPC_SERVER_STARTED", "Grpc server started in server with endpoint : "+edgeServer.ServerAddr)
	go grpcServer.Serve(lis)
}
