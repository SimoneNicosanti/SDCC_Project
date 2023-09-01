package server

import (
	"edge/proto/client"
)

type FileServiceServer struct {
	client.UnimplementedFileServiceServer
}

func checkTicket(requestId string) int {
	for index, authRequestId := range authorizedTicketIDs.IDs {
		if requestId == authRequestId {
			return index
		}
	}
	return -1
}
