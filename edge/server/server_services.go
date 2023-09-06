package server

import (
	"crypto/sha256"
	"edge/cache"
	"edge/peer"
	"edge/proto/client"
	"edge/s3_boundary"
	"edge/utils"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"syscall"

	"golang.org/x/net/context"

	"google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	status "google.golang.org/grpc/status"
)

func (s *FileServiceServer) Upload(uploadStream client.FileService_UploadServer) error {
	// Apri il file locale dove verranno scritti i chunks
	md, thereIsMetadata := metadata.FromIncomingContext(uploadStream.Context())
	if !thereIsMetadata {
		log.Println("NO METADATA")
		return status.Error(codes.Code(client.ErrorCodes_INVALID_TICKET), "[*ERROR*] - Request No Ticket")
	}
	ticketID := md.Get("ticket_id")[0]
	file_name := md.Get("file_name")[0]
	file_size, err := strconv.Atoi(md.Get("file_size")[0])
	if err != nil {
		log.Println("Impossibile effettuare il cast della size. SIZE = " + md.Get("file_size")[0])
	}

	// TODO Vedere come gestire grandezza dei file --> Passiamo dimensione del file nel context

	isValidRequest := checkTicket(ticketID)
	if isValidRequest == -1 {
		return status.Error(codes.Code(client.ErrorCodes_INVALID_TICKET), "[*ERROR*] - Request with Invalid Ticket")
	}
	defer publishNewTicket(isValidRequest)

	// Salvo prima su file locale o su S3?? PRIMA S3
	fileChannel := make(chan []byte, utils.GetIntegerEnvironmentVariable("UPLOAD_CHANNEL_SIZE"))
	defer close(fileChannel)

	uploadStreamReader := s3_boundary.UploadStream{ClientStream: uploadStream, FileName: file_name, FileChannel: fileChannel, ResidualChunk: make([]byte, 0)}

	// Thread in attesa di ricevere il file durante l'invio ad S3 in modo da salvarlo localmente
	go cache.GetCache().InsertFileInCache(fileChannel, file_name, file_size)
	//go cache.WriteChunksOnFile(fileChannel, file_name)
	err = s3_boundary.SendToS3(file_name, uploadStreamReader)
	if err != nil {
		utils.PrintEvent("UPLOAD_ERROR", fmt.Sprintf("Errore nel caricare il file '%s'\r\nTICKET: '%s'", file_name, ticketID))
		return status.Error(codes.Code(client.ErrorCodes_S3_ERROR), "[*ERROR*] - File Upload to S3 encountered some error")
	}
	utils.PrintEvent("UPLOAD_SUCCESS", fmt.Sprintf("File '%s' caricato con successo\r\nTICKET: '%s'", file_name, ticketID))
	response := client.Response{TicketId: ticketID, Success: true}
	err = uploadStream.SendAndClose(&response)
	if err != nil {
		return status.Error(codes.Code(client.ErrorCodes_CHUNK_ERROR), "[*ERROR*] - Couldn't close clientstream")
	}

	return nil
}

// TODO impostare un timer dopo il quale assumo che il file non sia stato trovato --> limito l'attesa
func (s *FileServiceServer) Download(requestMessage *client.FileDownloadRequest, downloadStream client.FileService_DownloadServer) error {
	utils.PrintEvent("CLIENT_REQUEST_RECEIVED", "Ricevuta richiesta di download per file '"+requestMessage.FileName+"'.")
	isValidRequest := checkTicket(requestMessage.TicketId)
	if isValidRequest == -1 {
		return status.Error(codes.Code(client.ErrorCodes_INVALID_TICKET), "[*ERROR*] - Invalid Ticket Request")
	}
	defer publishNewTicket(isValidRequest)

	_, err := os.Stat("/files/" + requestMessage.FileName)
	if os.IsNotExist(err) {
		fileRequest := peer.FileRequestMessage{FileName: requestMessage.FileName, TTL: utils.GetIntegerEnvironmentVariable("REQUEST_TTL"), TicketId: requestMessage.TicketId, SenderPeer: peer.SelfPeer, ForwarderPeer: peer.SelfPeer}
		ownerEdge, err := peer.NeighboursFileLookup(fileRequest)

		fileChannel := make(chan []byte, utils.GetIntegerEnvironmentVariable("DOWNLOAD_CHANNEL_SIZE"))
		defer close(fileChannel)
		// TODO Modificare la risposta in modo che torni la size del file!!
		go cache.GetCache().InsertFileInCache(fileChannel, requestMessage.FileName, 0)

		if err == nil { // ce l'ha ownerEdge --> Ricevi file come stream e invia chunk (+ salva in locale)
			utils.PrintEvent("FILE_IN_NETWORK", fmt.Sprintf("Il file '%s' è stato trovato nell'edge %s", requestMessage.FileName, ownerEdge.IpAddr))
			sendFromOtherEdge(ownerEdge, requestMessage, downloadStream, fileChannel)
		} else { // ce l'ha S3 --> Ricevi file come stream e invia chunk (+ salva in locale)
			// log.Printf("[*S3*] -> looking for file '%s' in s3...", requestMessage.FileName)
			// s3DownloadStream := s3_boundary.DownloadStream{ClientStream: downloadStream, FileName: requestMessage.FileName, TicketID: requestMessage.TicketId, FileChannel: fileChannel}
			// err = s3_boundary.SendFromS3(requestMessage, s3DownloadStream)
			// if err != nil {
			// 	errorHash := sha256.Sum256([]byte("[*ERROR*]"))
			// 	fileChannel <- errorHash[:]
			// 	return status.Error(codes.Code(client.ErrorCodes_FILE_NOT_FOUND_ERROR), "[*ERROR*] -> Couldn't locate requested file in specified bucket")
			// }
			return status.Error(codes.Code(client.ErrorCodes_FILE_NOT_FOUND_ERROR), "[*ERROR*] -> S3 DISABILITATO")
		}
	} else if err == nil { // ce l'ha l'edge corrente --> Leggi file e invia chunk
		utils.PrintEvent("CACHE", fmt.Sprintf("Il file '%s' è stato trovato nella cache", requestMessage.FileName))
		return sendFromLocalFileStream(requestMessage, downloadStream)
	} else { //Got an error
		// TODO Aggiungere errore per path sbagliato
		return err
	}
	utils.PrintEvent("DOWNLOAD_SUCCESS", fmt.Sprintf("Invio del file '%s' Completata", requestMessage.FileName))
	return nil
}

func sendFromOtherEdge(ownerEdge peer.PeerFileServer, requestMessage *client.FileDownloadRequest, clientDownloadStream client.FileService_DownloadServer, fileChannel chan []byte) error {
	// 1] Open gRPC connection to ownerEdge
	// 2] retrieve chunk by chunk (send to client + save in local)
	conn, err := grpc.Dial(ownerEdge.IpAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		errorHash := sha256.Sum256([]byte("[*ERROR*]"))
		fileChannel <- errorHash[:]
		return status.Error(codes.Code(client.ErrorCodes_STREAM_CLOSE_ERROR), "[*ERROR*] - Failed while trying to Dial edge via gRPC")
	}
	grpcClient := client.NewEdgeFileServiceClient(conn)
	edgeDownloadStream, err := grpcClient.DownloadFromEdge(context.Background(), &client.FileDownloadRequest{TicketId: "", FileName: requestMessage.FileName})
	if err != nil {
		errorHash := sha256.Sum256([]byte("[*ERROR*]"))
		fileChannel <- errorHash[:]
		return status.Error(codes.Code(client.ErrorCodes_STREAM_CLOSE_ERROR), "[*ERROR*] - Failed while trying to setup edge download stream via gRPC")
	}

	for {
		edgeChunk, err := edgeDownloadStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			errorHash := sha256.Sum256([]byte("[*ERROR*]"))
			fileChannel <- errorHash[:]
			return status.Error(codes.Code(client.ErrorCodes_CHUNK_ERROR), "[*ERROR*] - Failed while receiving chunks from other edge via gRPC")
		}
		clientDownloadStream.Send(&client.FileChunk{Chunk: edgeChunk.Chunk})
		fileChannel <- edgeChunk.Chunk
	}

	edgeDownloadStream.CloseSend()

	return nil
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
		downloadStream.Send(&client.FileChunk{Chunk: buffer[:n]})
	}
	return nil
}
