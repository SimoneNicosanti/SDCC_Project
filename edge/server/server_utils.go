package server

import (
	"crypto/sha256"
	"edge/proto/client"
	"edge/proto/edge"
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

// TODO Potrebbero esserci dei problemi di blocco nel caso in cui il thread consumer vada in errore: il channel si satura e non si procede
func writeChunksOnFile(fileChannel chan []byte, fileName string) error {
	localFile, err := os.Create("/files/" + fileName)
	if err != nil {
		return status.Error(codes.Code(edge.ErrorCodes_FILE_CREATE_ERROR), "[*ERROR*] - File creation failed")
	}
	syscall.Flock(int(localFile.Fd()), syscall.F_WRLCK)
	defer syscall.Flock(int(localFile.Fd()), syscall.F_UNLCK)
	defer localFile.Close()

	errorHashString := fmt.Sprintf("%x", sha256.Sum256([]byte("[*ERROR*]")))
	for chunk := range fileChannel {
		chunkString := fmt.Sprintf("%x", chunk)
		if strings.Compare(chunkString, errorHashString) == 0 {
			log.Println("[*ABORT*] -> Error occurred, removing file...")
			return os.Remove("/files/" + fileName)
		}
		_, err = localFile.Write(chunk)
		if err != nil {
			os.Remove("/files/" + fileName)
			return status.Error(codes.Code(edge.ErrorCodes_FILE_WRITE_ERROR), "[*ERROR*] - Couldn't write chunk on local file")
		}
	}
	log.Printf("[*SUCCESS*] - File '%s' caricato localmente con successo\r\n", fileName)

	return nil
}

func checkTicket(requestId string) int {
	for index, authRequestId := range authorizedTicketIDs.IDs {
		if requestId == authRequestId {
			return index
		}
	}
	return -1
}
