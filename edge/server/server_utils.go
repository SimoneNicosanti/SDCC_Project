package server

import (
	"edge/proto/file_transfer"
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
	file_transfer.UnimplementedFileServiceServer
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
