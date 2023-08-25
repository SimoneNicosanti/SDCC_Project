package server

import (
	"edge/proto"
	"edge/utils"
	"fmt"
	"io"
	"os"
)

type FileServiceServer struct {
	proto.UnimplementedFileServiceServer
}

func (s *FileServiceServer) Upload(uploadStream proto.FileService_UploadServer) error {
	// Apri il file locale dove verranno scritti i chunks
	message, err := uploadStream.Recv()
	if err != nil {
		return fmt.Errorf("[*ERROR*] - Failed while receiving chunks from clientstream via gRPC\n%s", err.Error())
	}

	localFile, err := os.Create("/files/" + message.FileName)
	if err != nil {
		return fmt.Errorf("[*ERROR*] - File creation failed\n%s", err.Error())
	}
	defer localFile.Close()

	for {
		_, err = localFile.Write(message.Chunk)
		if err != nil {
			return fmt.Errorf("[*ERROR*] - Couldn't write chunk on local file\r\n%s", err.Error())
		}

		message, err = uploadStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("[*ERROR*] - Failed while receiving chunks from clientstream via gRPC\n%s", err.Error())
		}
	}

	response := proto.Response{RequestId: message.RequestId, Success: true}
	err = uploadStream.SendAndClose(&response)
	if err != nil {
		return fmt.Errorf("[*ERROR*] - Couldn't close clientstream\r\n%s", err.Error())
	}
	// if response.RequestId != message.RequestId {
	// 	log.Printf("[*ERROR*] - RequestID '%d' non riconosciuto! Expected --> '%d' \r\n", response.RequestId, message.RequestId)
	// } else if !response.Success {
	// 	log.Printf("[*ERROR*] - nello scaricamento del File '%s' [REQ_ID: %d]\r\n", message.FileName, message.RequestId)
	// } else {
	// 	log.Printf("[*SUCCESS*] - File '%s' caricato con successo [REQ_ID: %d]\r\n", message.FileName, message.RequestId)
	// }

	return nil
}

func (s *FileServiceServer) Download(requestMessage *proto.FileDownloadRequest, downloadStream proto.FileService_DownloadServer) error {
	localFile, err := os.Open("/files/" + requestMessage.FileName)
	if err != nil {
		return fmt.Errorf("[*ERROR*] - File open failed\n%s", err.Error())
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
			return fmt.Errorf("[*ERROR*] - Failed during read operation\r\n%s", err.Error())
		}
		downloadStream.Send(&proto.FileChunk{RequestId: requestMessage.RequestId, FileName: requestMessage.FileName, Chunk: buffer[:n]})
	}

	return nil
}
