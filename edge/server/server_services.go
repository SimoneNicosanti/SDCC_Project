package server

import (
	"context"
	"crypto/sha256"
	"edge/peer"
	"edge/proto/client"
	"edge/proto/edge"
	"edge/utils"
	"fmt"
	"io"
	"log"
	"os"
	"syscall"

	"google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	status "google.golang.org/grpc/status"
)

type FileServiceServer struct {
	client.UnimplementedFileServiceServer
}

func (s *FileServiceServer) Upload(uploadStream client.FileService_UploadServer) error {
	// Apri il file locale dove verranno scritti i chunks

	message, err := uploadStream.Recv()
	if err != nil {
		return status.Error(codes.Code(client.ErrorCodes_CHUNK_ERROR), "[*ERROR*] - Failed while receiving chunks from clientstream via gRPC")
	}

	isValidRequest := checkTicket(message.TicketId)
	if isValidRequest == -1 {
		return status.Error(codes.Code(client.ErrorCodes_INVALID_TICKET), "[*ERROR*] - Request with Invalid Ticket")
	}
	defer publishNewTicket(isValidRequest)

	// TODO Aggiungere logica di aggiunta file ad S3
	localFile, err := os.Create("/files/" + message.FileName)
	if err != nil {
		return status.Error(codes.Code(client.ErrorCodes_FILE_CREATE_ERROR), "[*ERROR*] - File creation failed")
	}
	defer localFile.Close()

	// Salvo prima su file locale o su S3?? PRIMA S3
	fileChannel := make(chan []byte)
	defer close(fileChannel)
	go writeFileOnS3(fileChannel, message.FileName)

	for {
		_, err = localFile.Write(message.Chunk)
		if err != nil {
			return status.Error(codes.Code(client.ErrorCodes_FILE_WRITE_ERROR), "[*ERROR*] - Couldn't write chunk on local file")
		}
		fileChannel <- message.Chunk

		message, err = uploadStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return status.Error(codes.Code(client.ErrorCodes_CHUNK_ERROR), "[*ERROR*] - Failed while receiving chunks from clientstream via gRPC")
		}
	}

	log.Printf("[*SUCCESS*] - File '%s' caricato con successo [REQ_ID: %s]\r\n", message.FileName, message.TicketId)
	response := client.Response{TicketId: message.TicketId, Success: true}
	err = uploadStream.SendAndClose(&response)
	if err != nil {
		return status.Error(codes.Code(client.ErrorCodes_CHUNK_ERROR), "[*ERROR*] - Couldn't close clientstream")
	}

	return nil
}

func writeFileOnS3(fileChannel chan []byte, fileName string) {
	for chunk := range fileChannel {
		log.Println(chunk)
	}
}

//TODO lancio N thread (N = #vicini) e aspetto solo risposta affermativa --> altrimenti timeout e contatto S3 --> OK->SEND_FILE/FILE_NOT_FOUND_ERROR
// per limitare il numero di thread lanciati lo richiedo in modulo alla threshold (chiedo ai primi k, poi ai secondi k, etc etc...)
// Se sono presenti vicini con un riscontro positivo sul loro filtro di bloom, considereremo loro nei primi k da contattare (eventualmente
// aggiungiamo altri casualmente se non arriviamo a k)
// imposto un timer dopo il quale assumo che il file non sia stato trovato --> limito l'attesa
// imposto un TTL per il numero di hop della richiesta (faccio una ricerca relativamente locale)
func (s *FileServiceServer) Download(requestMessage *client.FileDownloadRequest, downloadStream client.FileService_DownloadServer) error {
	isValidRequest := checkTicket(requestMessage.TicketId)
	if isValidRequest == -1 {
		return status.Error(codes.Code(client.ErrorCodes_INVALID_TICKET), "[*ERROR*] - Invalid Ticket Request")
	}
	defer publishNewTicket(isValidRequest)

	_, err := os.Stat("/files/" + requestMessage.FileName)
	if os.IsNotExist(err) {
		fileRequest := peer.FileRequestMessage{FileName: requestMessage.FileName, TTL: utils.GetIntegerEnvironmentVariable("REQUEST_TTL")}
		ownerEdge, err := peer.NeighboursFileLookup(fileRequest)
		if err == nil {
			// ce l'ha ownerEdge --> Ricevi file come stream e invia chunk (+ salva in locale)
			sendFromOtherEdge(ownerEdge, requestMessage, downloadStream)
		} else {
			// ce l'ha S3 --> Ricevi file come stream e invia chunk (+ salva in locale)
			sendFromS3()
		}
	} else if err == nil {
		// ce l'ho io --> Leggi file e invia chunk
		return sendFromLocalFileStream(requestMessage, downloadStream)
	} else { //Got an error
		// TODO Aggiungere errore per path sbagliato

	}

	return nil
}

func sendFromOtherEdge(ownerEdge peer.EdgePeer, requestMessage *client.FileDownloadRequest, clientDownloadStream client.FileService_DownloadServer) error {
	// 1] Open gRPC connection to ownerEdge
	// 2] retrieve chunk by chunk (send to client + save in local)
	conn, err := grpc.Dial(ownerEdge.PeerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return status.Error(codes.Code(edge.ErrorCodes_STREAM_CLOSE_ERROR), "[*ERROR*] - Failed while trying to Dial edge via gRPC")
	}
	grpcClient := edge.NewEdgeFileServiceClient(conn)
	edgeDownloadStream, err := grpcClient.DownloadFromEdge(context.Background(), &edge.EdgeFileDownloadRequest{})
	if err != nil {
		return status.Error(codes.Code(edge.ErrorCodes_STREAM_CLOSE_ERROR), "[*ERROR*] - Failed while trying to setup edge download stream via gRPC")
	}
	fileChannel := make(chan []byte)
	defer close(fileChannel)
	go writeChunksOnFile(fileChannel, requestMessage.FileName)
	for {
		edgeChunk, err := edgeDownloadStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			errorHash := sha256.Sum256([]byte("[*ERROR*]"))
			fileChannel <- errorHash[:]
			return status.Error(codes.Code(edge.ErrorCodes_CHUNK_ERROR), "[*ERROR*] - Failed while receiving chunks from other edge via gRPC")
		}
		clientDownloadStream.Send(&client.FileChunk{TicketId: requestMessage.TicketId, FileName: requestMessage.FileName, Chunk: edgeChunk.Chunk})
		fileChannel <- edgeChunk.Chunk
	}

	edgeDownloadStream.CloseSend()

	return nil
}

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
		chunkString := string(chunk)
		if chunkString == errorHashString {
			return os.Remove("/files/" + fileName)
		}
		_, err = localFile.Write(chunk)
		if err != nil {
			os.Remove("/files/" + fileName)
			return status.Error(codes.Code(edge.ErrorCodes_FILE_WRITE_ERROR), "[*ERROR*] - Couldn't write chunk on local file")
		}
	}
	log.Printf("[*LOAD_SUCCESS*] - File '%s' caricato localmente con successo\r\n", fileName)

	return nil
}

func sendFromS3() {
	// 1] Open connection to S3
	// 2] retrieve chunk by chunk (send to client + save in local)
}

func sendFromLocalFileStream(requestMessage *client.FileDownloadRequest, downloadStream client.FileService_DownloadServer) error {
	localFile, err := os.Open("/files/" + requestMessage.FileName)
	if err != nil {
		return status.Error(codes.Code(client.ErrorCodes_FILE_NOT_FOUND_ERROR), "[*ERROR*] - File opening failed")
	}
	syscall.Flock(int(localFile.Fd()), syscall.F_RDLCK)
	defer syscall.Flock(int(localFile.Fd()), syscall.F_UNLCK)
	defer localFile.Close()

	chunkSize := utils.GetIntegerEnvironmentVariable("CHUNK_SIZE")
	buffer := make([]byte, chunkSize)
	for {
		n, err := localFile.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return status.Error(codes.Code(client.ErrorCodes_FILE_READ_ERROR), "[*ERROR*] - Failed during read operation\r")
		}
		downloadStream.Send(&client.FileChunk{TicketId: requestMessage.TicketId, FileName: requestMessage.FileName, Chunk: buffer[:n]})
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

/*
	go func() {
		channel int
		for {
			conn := Accept()
			channel <- 0
			go Serve(conn)
			defer {
				<- channel
				close(conn)
			}
		}
	}
*/
