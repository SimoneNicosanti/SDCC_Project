package server

import (
	"edge/proto/client"
	"sync"
)

type FileServiceServer struct {
	client.UnimplementedFileServiceServer
	currentServerLoad int
	mutex             sync.RWMutex
}

func checkTicket(requestId string) int {
	for index, authRequestId := range authorizedTicketIDs.IDs {
		if requestId == authRequestId {
			return index
		}
	}
	return -1
}

func (server *FileServiceServer) incrementWorkload() {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	server.currentServerLoad++
}

func (server *FileServiceServer) decrementWorkload() {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	server.currentServerLoad--
}
