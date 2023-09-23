package services

import (
	"context"
	"fmt"
	"load_balancer/login"
	proto "load_balancer/proto/load_balancer"
	"load_balancer/utils"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var balancingServer BalancingServiceServer = BalancingServiceServer{
	mapMutex:           sync.RWMutex{},
	edgeServerMap:      map[EdgeServer]int{},
	heartbeatCheckTime: time.Now(),
	heartbeats:         map[EdgeServer](time.Time){},
	sequenceNumber:     0,
}

func (balancingServer *BalancingServiceServer) LoginClient(ctx context.Context, userInfo *proto.User) (*proto.LoginResponse, error) {
	utils.PrintEvent("CLIENT_LOGIN_ATTEMPT", fmt.Sprintf("Il Client '%s' sta effettuando il log in...", userInfo.Username))
	return &proto.LoginResponse{Logged: login.UserLogin(userInfo.Username, userInfo.Passwd)}, nil
}

func (balancingServer *BalancingServiceServer) GetEdge(ctx context.Context, userInfo *proto.User) (*proto.BalancerResponse, error) {
	defer convertAndPrintEdgeServerMap(balancingServer.edgeServerMap)
	utils.PrintEvent("GET_EDGE_SERVER", fmt.Sprintf("Richiesta ricevuta da '%s', selezione dell'edge con carico minimo in corso...", userInfo.Username))
	success := login.UserLogin(userInfo.Username, userInfo.Passwd)
	var edgeIpAddr string
	var err error
	if success {
		edgeIpAddr, err = balancingServer.pickEdgeServer()
		if err != nil {
			return &proto.BalancerResponse{Success: false, EdgeIpAddr: ""}, status.Error(codes.Unauthenticated, err.Error())
		}
	} else {
		edgeIpAddr = ""
	}
	newRequestId := utils.ConvertToString(balancingServer.sequenceNumber)
	balancingServer.sequenceNumber++
	return &proto.BalancerResponse{Success: success, EdgeIpAddr: edgeIpAddr, RequestId: newRequestId}, nil
}

func (balancingServer *BalancingServiceServer) NotifyJobEnd(edgeServer EdgeServer, returnPtr *int) error {
	defer convertAndPrintEdgeServerMap(balancingServer.edgeServerMap)
	utils.PrintEvent("EDGE_SERVER_JOB_END", fmt.Sprintf("L'Edge Server '%s' ha completato il job", edgeServer.ServerAddr))
	*returnPtr = 0

	balancingServer.mapMutex.Lock()
	defer balancingServer.mapMutex.Unlock()

	value, isInMap := balancingServer.edgeServerMap[edgeServer]
	if isInMap && value > 0 {
		balancingServer.edgeServerMap[edgeServer]--
	}
	return nil
}

func (balancingServer *BalancingServiceServer) Heartbeat(heartbeatMessage HeartbeatMessage, replyPtr *int) error {
	*replyPtr = 0

	balancingServer.mapMutex.Lock()
	defer balancingServer.mapMutex.Unlock()

	edgeServer := heartbeatMessage.EdgeServer
	_, isInMap := balancingServer.heartbeats[edgeServer]
	if !isInMap {
		utils.PrintEvent("ACTIVE_EDGE_SERVER_FOUND", fmt.Sprintf("Edge Server '%s' è attivo!", edgeServer.ServerAddr))
		balancingServer.edgeServerMap[edgeServer] = heartbeatMessage.CurrentLoad
		convertAndPrintEdgeServerMap(balancingServer.edgeServerMap)
	}

	balancingServer.heartbeats[edgeServer] = time.Now()

	return nil
}
