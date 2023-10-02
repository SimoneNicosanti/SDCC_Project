package services

import (
	proto "load_balancer/proto/load_balancer"
	"sync"
	"time"
)

type EdgeServer struct {
	ServerAddr string
}

type HeartbeatMessage struct {
	EdgeServer  EdgeServer
	CurrentLoad int
}

type BalancingServiceServer struct {
	proto.UnimplementedBalancingServiceServer
	mapMutex           sync.RWMutex
	edgeServerMap      map[EdgeServer]int
	heartbeatCheckTime time.Time
	heartbeats         map[EdgeServer](time.Time)
	sequenceMutex      sync.RWMutex
	sequenceNumber     int64
}
