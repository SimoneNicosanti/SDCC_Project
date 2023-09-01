package server

import (
	"crypto/sha256"
	"edge/proto/client"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"

	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
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
