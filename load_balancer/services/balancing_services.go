package services

import (
	"context"
	"fmt"
	"load_balancer/login"
	proto "load_balancer/proto/load_balancer"
	"load_balancer/utils"
	"math"
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
	utils.PrintEvent("CLIENT_LOGIN_ATTEMPT", fmt.Sprintf("Client '%s' is trying to Log in...", userInfo.Username))
	return &proto.LoginResponse{Logged: login.UserLogin(userInfo.Username, userInfo.Passwd)}, nil
}

func (balancingServer *BalancingServiceServer) GetEdge(ctx context.Context, userInfo *proto.User) (*proto.BalancerResponse, error) {
	utils.PrintEvent("GET_EDGE_SERVER", fmt.Sprintf("Request from user '%s', taking edge server with minimum load", userInfo.Username))
	success := login.UserLogin(userInfo.Username, userInfo.Passwd)
	var edgeIpAddr string
	var err error
	if success {
		edgeIpAddr, err = balancingServer.PickEdgeServer()
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

func (balancingServer *BalancingServiceServer) NotifyJobEnd(edgeServer EdgeServer, returnPtr *int) error {
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
