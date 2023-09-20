package server

import (
	"edge/proto/client"
	"sync"
)

type HeartbeatMessage struct {
	EdgeServer  EdgeServer
	CurrentLoad int
}

type EdgeServer struct {
	ServerAddr string
}

type FileServiceServer struct {
	client.UnimplementedFileServiceServer
	currentWorkload int
	mutex           sync.RWMutex
}

func (server *FileServiceServer) incrementWorkload() {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	server.currentWorkload++
}

func (server *FileServiceServer) decrementWorkload() {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	server.currentWorkload--
}
