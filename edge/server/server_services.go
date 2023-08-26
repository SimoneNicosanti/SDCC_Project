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

	//TODO Add ticket release on rabbitMQ Queue

	return nil
}

func (s *FileServiceServer) Download(requestMessage *proto.FileDownloadRequest, downloadStream proto.FileService_DownloadServer) error {
	isValidRequest := checkTicket(requestMessage.TicketId)
	if isValidRequest == -1 {
		return status.Error(codes.Code(proto.ErrorCodes_INVALID_TICKET), "[*ERROR*] - Invalid Ticket Request")
	}
	defer publishNewTicket(isValidRequest)
	localFile, err := os.Open("/files/" + requestMessage.FileName)
	if err != nil {
		return status.Error(codes.Code(proto.ErrorCodes_FILE_NOT_FOUND_ERROR), "[*ERROR*] - File opening failed")
	}
	defer localFile.Close()
	chunkSize := utils.GetIntegerEnvironmentVariable("CHUNK_SIZE") // dimensione del chunk
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

	publishNewTicket(isValidRequest)
	//TODO Add ticket release on rabbitMQ Queue
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
