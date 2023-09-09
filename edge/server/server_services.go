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
	file_size, err := strconv.ParseInt(md.Get("file_size")[0], 10, 64)
	if err != nil {
		utils.PrintEvent("CAST_ERR", "Impossibile effettuare il cast della size. SIZE = "+md.Get("file_size")[0])
	}
	utils.PrintEvent("CLIENT_REQUEST_RECEIVED", "Ricevuta richiesta di upload per file '"+file_name+"'.")

	isValidRequest := checkTicket(ticketID)
	if isValidRequest == -1 {
		return status.Error(codes.Code(client.ErrorCodes_INVALID_TICKET), "[*ERROR*] - Request with Invalid Ticket")
	}
	defer publishNewTicket(isValidRequest)

	// Salvo prima su S3 e poi su file locale
	fileChannel := make(chan []byte, utils.GetIntEnvironmentVariable("UPLOAD_CHANNEL_SIZE"))
	defer close(fileChannel)

	isFileCacheable := cache.CheckFileSize(file_size)
	uploadStreamReader := s3_boundary.UploadStream{ClientStream: uploadStream, FileName: file_name, FileChannel: fileChannel, ResidualChunk: make([]byte, 0), WriteOnChannel: isFileCacheable}

	if isFileCacheable {
		// Thread in attesa di ricevere il file durante l'invio ad S3 in modo da salvarlo localmente
		go cache.GetCache().InsertFileInCache(fileChannel, file_name, file_size)
		// go cache.WriteChunksInCache(fileChannel, file_name)
	}

	err = s3_boundary.SendToS3(file_name, uploadStreamReader)
	if err != nil {
		utils.PrintEvent("UPLOAD_ERROR", fmt.Sprintf("Errore nel caricare il file '%s'\r\nTICKET: '%s'", file_name, ticketID))
		return status.Error(codes.Code(client.ErrorCodes_S3_ERROR), fmt.Sprintf("[*ERROR*] - File Upload to S3 encountered some error.\r\nError: '%s'", err.Error()))
	}
	utils.PrintEvent("UPLOAD_SUCCESS", fmt.Sprintf("File '%s' caricato con successo\r\nTICKET: '%s'", file_name, ticketID))
	response := client.Response{TicketId: ticketID, Success: true}
	err = uploadStream.SendAndClose(&response)
	if err != nil {
		return status.Error(codes.Code(client.ErrorCodes_CHUNK_ERROR), fmt.Sprintf("[*ERROR*] - Couldn't close clientstream.\r\nError: '%s'", err.Error()))
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

	if !cache.GetCache().IsFileInCache(requestMessage.FileName) {
		fileRequest := peer.FileRequestMessage{FileName: requestMessage.FileName, TTL: utils.GetIntEnvironmentVariable("REQUEST_TTL"), TicketId: requestMessage.TicketId, SenderPeer: peer.SelfPeer}
		lookupReponse, err := peer.NeighboursFileLookup(fileRequest)

		var ownerEdge peer.PeerFileServer
		var fileSize int64
		var isFileCacheable bool

		fileChannel := make(chan []byte, utils.GetIntEnvironmentVariable("DOWNLOAD_CHANNEL_SIZE"))
		defer close(fileChannel)
		if err == nil {
			ownerEdge = lookupReponse.OwnerEdge
			fileSize = lookupReponse.FileSize
			isFileCacheable = cache.CheckFileSize(fileSize)
			// TODO Modificare la risposta in modo che torni la size del file!!
			if isFileCacheable {
				// Thread in attesa di ricevere il file durante l'invio ad S3 in modo da salvarlo localmente
				go cache.GetCache().InsertFileInCache(fileChannel, requestMessage.FileName, fileSize)
			}
		}

		if err == nil { // ce l'ha ownerEdge --> Ricevi file come stream e invia chunk (+ salva in locale)
			utils.PrintEvent("FILE_IN_NETWORK", fmt.Sprintf("Il file '%s' è stato trovato nell'edge %s", requestMessage.FileName, ownerEdge.IpAddr))
			err := sendFromOtherEdge(ownerEdge, requestMessage, downloadStream, fileChannel, isFileCacheable)
			if err != nil {
				utils.PrintEvent("SEND_FROM_EDGE_ERROR", err.Error())
			}
		} else { // ce l'ha S3 --> Ricevi file come stream e invia chunk (+ salva in locale)
			// log.Printf("[*S3*] -> looking for file '%s' in s3...", requestMessage.FileName)
			// s3DownloadStream := s3_boundary.DownloadStream{ClientStream: downloadStream, FileName: requestMessage.FileName, TicketID: requestMessage.TicketId, FileChannel: fileChannel, WriteOnChannel: isFileCacheable}
			// err = s3_boundary.SendFromS3(requestMessage, s3DownloadStream)
			// if err != nil {
			// 	errorHash := sha256.Sum256([]byte("[*ERROR*]"))
			// 	fileChannel <- errorHash[:]
			// 	return status.Error(codes.Code(client.ErrorCodes_FILE_NOT_FOUND_ERROR), fmt.Sprintf("[*ERROR*] -> Couldn't locate requested file in specified bucket.\r\nError: '%s'", err.Error()))
			// }
			return status.Error(codes.Code(client.ErrorCodes_FILE_NOT_FOUND_ERROR), "[*ERROR*] -> S3 DISABILITATO")
		}
	} else { // ce l'ha l'edge corrente --> Leggi file e invia chunk
		utils.PrintEvent("CACHE", fmt.Sprintf("Il file '%s' è stato trovato nella cache", requestMessage.FileName))
		return sendFromLocalFileStream(requestMessage, downloadStream)
	}
	utils.PrintEvent("DOWNLOAD_SUCCESS", fmt.Sprintf("Invio del file '%s' Completata", requestMessage.FileName))
	return nil
}

func sendFromOtherEdge(ownerEdge peer.PeerFileServer, requestMessage *client.FileDownloadRequest, clientDownloadStream client.FileService_DownloadServer, fileChannel chan []byte, writeOnChannel bool) error {
	// 1] Open gRPC connection to ownerEdge
	// 2] retrieve chunk by chunk (send to client + save in local)
	opts := []grpc.DialOption{
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(utils.GetIntEnvironmentVariable("MAX_GRPC_MESSAGE_SIZE"))), // Imposta la nuova dimensione massima
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	conn, err := grpc.Dial(ownerEdge.IpAddr, opts...)
	if err != nil {
		errorHash := sha256.Sum256([]byte("[*ERROR*]"))
		fileChannel <- errorHash[:]
		return status.Error(codes.Code(client.ErrorCodes_STREAM_CLOSE_ERROR), fmt.Sprintf("[*ERROR*] - Failed while trying to dial ownerEdge via gRPC.\r\nError: '%s'", err.Error()))
	}
	grpcClient := client.NewEdgeFileServiceClient(conn)
	context := context.Background()
	edgeDownloadStream, err := grpcClient.DownloadFromEdge(context, &client.FileDownloadRequest{TicketId: "", FileName: requestMessage.FileName})
	if err != nil {
		errorHash := sha256.Sum256([]byte("[*ERROR*]"))
		fileChannel <- errorHash[:]
		return status.Error(codes.Code(client.ErrorCodes_STREAM_CLOSE_ERROR), fmt.Sprintf("[*ERROR*] - Failed while triggering download from edge via gRPC.\r\nError: '%s'", err.Error()))
	}

	for {
		edgeChunk, err := edgeDownloadStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			errorHash := sha256.Sum256([]byte("[*ERROR*]"))
			fileChannel <- errorHash[:]
			return status.Error(codes.Code(client.ErrorCodes_CHUNK_ERROR), fmt.Sprintf("[*ERROR*] - Failed while receiving chunks from other edge via gRPC.\r\nError: '%s'", err.Error()))
		}
		clientDownloadStream.Send(&client.FileChunk{Chunk: edgeChunk.Chunk})

		if writeOnChannel {
			chunkCopy := make([]byte, len(edgeChunk.Chunk))
			copy(chunkCopy, edgeChunk.Chunk)
			fileChannel <- chunkCopy
		}
	}

	edgeDownloadStream.CloseSend()

	return nil
}

func sendFromLocalFileStream(requestMessage *client.FileDownloadRequest, downloadStream client.FileService_DownloadServer) error {
	localFile, err := os.Open(utils.GetEnvironmentVariable("FILES_PATH") + requestMessage.FileName)
	if err != nil {
		return status.Error(codes.Code(client.ErrorCodes_FILE_NOT_FOUND_ERROR), fmt.Sprintf("[*ERROR*] - File opening failed.\r\nError: '%s'", err.Error()))
	}
	syscall.Flock(int(localFile.Fd()), syscall.F_RDLCK)
	defer syscall.Flock(int(localFile.Fd()), syscall.F_UNLCK)
	defer localFile.Close()

	chunkSize := utils.GetIntEnvironmentVariable("CHUNK_SIZE")
	buffer := make([]byte, chunkSize)
	for {
		n, err := localFile.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return status.Error(codes.Code(client.ErrorCodes_FILE_READ_ERROR), fmt.Sprintf("[*ERROR*] - Failed during read operation.\r\nError: '%s'", err.Error()))
		}
		downloadStream.Send(&client.FileChunk{Chunk: buffer[:n]})
	}
	return nil
}
