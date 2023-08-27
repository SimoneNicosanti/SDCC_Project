package server

import (
	"edge/proto"
	"edge/utils"
	"io"
	"log"
	"os"

	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

type FileServiceServer struct {
	proto.UnimplementedFileServiceServer
}

func (s *FileServiceServer) Upload(uploadStream proto.FileService_UploadServer) error {
	// Apri il file locale dove verranno scritti i chunks

	message, err := uploadStream.Recv()
	if err != nil {
		return status.Error(codes.Code(proto.ErrorCodes_CHUNK_ERROR), "[*ERROR*] - Failed while receiving chunks from clientstream via gRPC")
	}

	isValidRequest := checkTicket(message.TicketId)
	if isValidRequest == -1 {
		return status.Error(codes.Code(proto.ErrorCodes_INVALID_TICKET), "[*ERROR*] - Request with Invalid Ticket")
	}
	defer publishNewTicket(isValidRequest)

	localFile, err := os.Create("/files/" + message.FileName)
	if err != nil {
		return status.Error(codes.Code(proto.ErrorCodes_FILE_CREATE_ERROR), "[*ERROR*] - File creation failed")
	}
	defer localFile.Close()

	for {
		_, err = localFile.Write(message.Chunk)
		if err != nil {
			return status.Error(codes.Code(proto.ErrorCodes_FILE_WRITE_ERROR), "[*ERROR*] - Couldn't write chunk on local file")
		}

		message, err = uploadStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return status.Error(codes.Code(proto.ErrorCodes_CHUNK_ERROR), "[*ERROR*] - Failed while receiving chunks from clientstream via gRPC")
		}
	}

	log.Printf("[*SUCCESS*] - File '%s' caricato con successo [REQ_ID: %s]\r\n", message.FileName, message.TicketId)
	response := proto.Response{TicketId: message.TicketId, Success: true}
	err = uploadStream.SendAndClose(&response)
	if err != nil {
		return status.Error(codes.Code(proto.ErrorCodes_CHUNK_ERROR), "[*ERROR*] - Couldn't close clientstream")
	}

	return nil
}

func (s *FileServiceServer) Download(requestMessage *proto.FileDownloadRequest, downloadStream proto.FileService_DownloadServer) error {
	isValidRequest := checkTicket(requestMessage.TicketId)
	if isValidRequest == -1 {
		return status.Error(codes.Code(proto.ErrorCodes_INVALID_TICKET), "[*ERROR*] - Invalid Ticket Request")
	}
	defer publishNewTicket(isValidRequest)
	//TODO lancio N thread (N = #vicini) e aspetto solo risposta affermativa --> altrimenti timeout e contatto S3 --> OK->SEND_FILE/FILE_NOT_FOUND_ERROR
	// per limitare il numero di thread lanciati lo richiedo in modulo alla threshold (chiedo ai primi k, poi ai secondi k, etc etc...)
	// Se sono presenti vicini con un riscontro positivo sul loro filtro di bloom, considereremo loro nei primi k da contattare (eventualmente
	// aggiungiamo altri casualmente se non arriviamo a k)
	// imposto un timer dopo il quale assumo che il file non sia stato trovato --> limito l'attesa
	// imposto un TTL per il numero di hop della richiesta (faccio una ricerca relativamente locale)
	_, err := os.Stat("/files/" + requestMessage.FileName)
	if os.IsNotExist(err) { //Non c'è nella mia memoria
		//TODO peer.go alla fine c'è gestione filtri di bloom (guarda primitiva)
		//TODO accedere alla struct dei neighbours
		return
	} else if err == nil { //L'ho trovato in memoria --> posso mandarlo direttamente io
		// dimensione del chunk
		return sendLocalFileStream("/files/"+requestMessage.FileName, requestMessage, downloadStream)
	} else { //Got an error
		return err
	}
}

func sendLocalFileStream(nameOfFile string, requestMessage *proto.FileDownloadRequest, downloadStream proto.FileService_DownloadServer) error {
	localFile, err := os.Open(nameOfFile)
	if err != nil {
		return status.Error(codes.Code(proto.ErrorCodes_FILE_NOT_FOUND_ERROR), "[*ERROR*] - File opening failed")
	}
	defer localFile.Close()
	chunkSize := utils.GetIntegerEnvironmentVariable("CHUNK_SIZE")
	buffer := make([]byte, chunkSize)
	for {
		n, err := localFile.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return status.Error(codes.Code(proto.ErrorCodes_FILE_READ_ERROR), "[*ERROR*] - Failed during read operation\r")
		}
		downloadStream.Send(&proto.FileChunk{TicketId: requestMessage.TicketId, FileName: requestMessage.FileName, Chunk: buffer[:n]})
	}
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
