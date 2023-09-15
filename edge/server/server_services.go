package server

import (
	"edge/cache"
	"edge/channels"
	"edge/peer"
	"edge/proto/client"
	"edge/s3_boundary"
	"edge/utils"
	"fmt"
	"io"
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
	ticketID, fileName, fileSize, shouldReturn, returnValue := retrieveMetadata(uploadStream)
	if shouldReturn {
		return returnValue
	}
	utils.PrintEvent("CLIENT_REQUEST_RECEIVED", fmt.Sprintf("Ricevuta richiesta di upload per file '%s'\r\nTicket: '%s'", fileName, ticketID))

	isValidRequest := checkTicket(ticketID)
	if isValidRequest == -1 {
		return status.Error(codes.Code(client.ErrorCodes_INVALID_TICKET), "[*ERROR*] - Request with Invalid Ticket")
	}
	defer publishNewTicket(isValidRequest)

	// Salvo prima su S3 e poi su file locale
	s3RedirectionChannel := channels.NewRedirectionChannel(utils.GetIntEnvironmentVariable("UPLOAD_CHANNEL_SIZE"))
	cacheRedirectionChannel := channels.NewRedirectionChannel(utils.GetIntEnvironmentVariable("UPLOAD_CHANNEL_SIZE"))

	// Ridirezione del flusso sulla cache
	isFileCacheable := cache.IsFileCacheable(fileSize)
	if isFileCacheable {
		// Thread in attesa di ricevere il file durante l'invio ad S3 in modo da salvarlo localmente
		go cache.GetCache().InsertFileInCache(cacheRedirectionChannel, fileName, fileSize)
	}

	// Ridirezione su S3
	go s3_boundary.SendToS3(fileName, s3RedirectionChannel)

	err := rcvAndRedirectChunks(s3RedirectionChannel, cacheRedirectionChannel, isFileCacheable, uploadStream)

	if err != nil {
		utils.PrintEvent("UPLOAD_ERROR", fmt.Sprintf("Errore nel caricare il file '%s'\r\nTICKET: '%s'", fileName, ticketID))
		return status.Error(codes.Code(client.ErrorCodes_S3_ERROR), fmt.Sprintf("[*ERROR*] - File Upload to S3 encountered some error.\r\nError: '%s'", err.Error()))
	}
	utils.PrintEvent("UPLOAD_SUCCESS", fmt.Sprintf("File '%s' caricato con successo\r\nTICKET: '%s'", fileName, ticketID))
	response := client.Response{TicketId: ticketID, Success: true}
	err = uploadStream.SendAndClose(&response)
	if err != nil {
		return status.Error(codes.Code(client.ErrorCodes_CHUNK_ERROR), fmt.Sprintf("[*ERROR*] - Impossibile chiudere il clientstream.\r\nError: '%s'", err.Error()))
	}

	return nil
}

func retrieveMetadata(uploadStream client.FileService_UploadServer) (string, string, int64, bool, error) {
	md, thereIsMetadata := metadata.FromIncomingContext(uploadStream.Context())
	if !thereIsMetadata {
		return "", "", 0, true, status.Error(codes.Code(client.ErrorCodes_INVALID_TICKET), "[*NO_METADATA*] - No metadata found")
	}
	ticketID := md.Get("ticket_id")[0]
	file_name := md.Get("file_name")[0]
	file_size, err := strconv.ParseInt(md.Get("file_size")[0], 10, 64)
	if err != nil {
		return "", "", 0, true, status.Error(codes.Code(client.ErrorCodes_INVALID_TICKET), fmt.Sprintf("[*CAST_ERROR*] - Impossibile effettuare il cast della size : '%s'", err.Error()))
	}
	return ticketID, file_name, file_size, false, nil
}

// TODO impostare un timer (sulla lookup) dopo il quale assumo che il file non sia stato trovato --> limito l'attesa
func (s *FileServiceServer) Download(requestMessage *client.FileDownloadRequest, downloadStream client.FileService_DownloadServer) error {
	utils.PrintEvent("CLIENT_REQUEST_RECEIVED", fmt.Sprintf("Ricevuta richiesta di download per file '%s'\r\nTicket: '%s'", requestMessage.FileName, requestMessage.TicketId))
	isValidRequest := checkTicket(requestMessage.TicketId)
	if isValidRequest == -1 {
		return status.Error(codes.Code(client.ErrorCodes_INVALID_TICKET), "[*ERROR*] - Invalid Ticket Request")
	}
	defer publishNewTicket(isValidRequest)

	if !cache.GetCache().IsFileInCache(requestMessage.FileName) {
		// Thread in attesa di ricevere il file durante l'invio ad S3 in modo da salvarlo localmente
		lookupReponse, err := lookupFileInNetwork(requestMessage)
		var askS3 bool = true
		if err == nil { // ce l'ha ownerEdge --> Ricevi file come stream e invia chunk (+ salva in locale)
			err = sendFromOtherEdge(lookupReponse, requestMessage.FileName, downloadStream)
			if err != nil {
				utils.PrintEvent("OTHEREDGE_DOWNLOAD_ERROR", fmt.Sprintf("Impossibile recuperare file '%s' da altro edge... Ripiego su S3\r\nError: '%s'", requestMessage.FileName, err.Error()))
			} else {
				askS3 = false
			}
		}
		if askS3 { // ce l'ha S3 --> Ricevi file come stream e invia chunk (+ salva in locale)
			err := redirectFromS3(requestMessage.FileName, downloadStream)
			if err != nil {
				utils.PrintEvent("S3_DOWNLOAD_ERROR", err.Error())
				return status.Error(codes.Code(client.ErrorCodes_FILE_NOT_FOUND_ERROR), fmt.Sprintf("[*ERROR*] -> Impossibile recuperare il file '%s' dal bucket specificato.\r\nError: '%s'", requestMessage.FileName, err.Error()))
			}
		}
	} else { // ce l'ha l'edge corrente --> Leggi file e invia chunk
		return sendFromLocalCache(requestMessage.FileName, downloadStream)
	}
	utils.PrintEvent("DOWNLOAD_SUCCESS", fmt.Sprintf("Invio del file '%s' Completata", requestMessage.FileName))
	return nil
}

func redirectFromS3(fileName string, downloadStream client.FileService_DownloadServer) error {
	utils.PrintEvent("S3_LOOKUP", fmt.Sprintf("Cercando il file '%s' in s3...", fileName))
	var isFileCachable bool = false
	fileSize, err := s3_boundary.GetFileSize(fileName)
	if err != nil {
		utils.PrintEvent("S3_ERROR", fmt.Sprintf("Impossibile ricavare dimensione del file '%s' da S3", fileName))
	} else {
		isFileCachable = cache.IsFileCacheable(fileSize)
	}

	clientRedirectionChannel := channels.NewRedirectionChannel(utils.GetIntEnvironmentVariable("DOWNLOAD_CHANNEL_SIZE"))
	cacheRedirectionChannel := channels.NewRedirectionChannel(utils.GetIntEnvironmentVariable("DOWNLOAD_CHANNEL_SIZE"))

	if isFileCachable {
		go cache.GetCache().InsertFileInCache(cacheRedirectionChannel, fileName, fileSize)
	}

	go s3_boundary.SendFromS3(fileName, clientRedirectionChannel, cacheRedirectionChannel, isFileCachable)

	err = redirectStreamToClient(clientRedirectionChannel, downloadStream)

	return err
}

func lookupFileInNetwork(requestMessage *client.FileDownloadRequest) (peer.FileLookupResponse, error) {
	fileRequest := peer.FileRequestMessage{FileName: requestMessage.FileName, TTL: utils.GetIntEnvironmentVariable("REQUEST_TTL"), TicketId: requestMessage.TicketId, SenderPeer: peer.SelfPeer}
	lookupReponse, err := peer.NeighboursFileLookup(fileRequest)

	if err != nil {
		return peer.FileLookupResponse{}, err
	}
	return lookupReponse, nil
}

func sendFromOtherEdge(lookupResponse peer.FileLookupResponse, fileName string, clientDownloadStream client.FileService_DownloadServer) error {
	utils.PrintEvent("FILE_IN_NETWORK", fmt.Sprintf("Il file '%s' è stato trovato nell'edge %s", fileName, lookupResponse.OwnerEdge.IpAddr))
	// 1] Open gRPC connection to ownerEdge
	// 2] retrieve chunk by chunk (send to client + save in local)
	clientRedirectionChannel := channels.NewRedirectionChannel(utils.GetIntEnvironmentVariable("DOWNLOAD_CHANNEL_SIZE"))
	cacheRedirectionChannel := channels.NewRedirectionChannel(utils.GetIntEnvironmentVariable("DOWNLOAD_CHANNEL_SIZE"))

	isFileCachable := cache.IsFileCacheable(lookupResponse.FileSize)
	if isFileCachable {
		go cache.GetCache().InsertFileInCache(cacheRedirectionChannel, fileName, lookupResponse.FileSize)
	}
	// Imposta la nuova dimensione massima
	go downloadFromOtherEdge(lookupResponse, fileName, cacheRedirectionChannel, clientRedirectionChannel, isFileCachable)

	return redirectStreamToClient(clientRedirectionChannel, clientDownloadStream)
}

func downloadFromOtherEdge(lookupResponse peer.FileLookupResponse, fileName string, cacheRedirectionChannel channels.RedirectionChannel, clientRedirectionChannel channels.RedirectionChannel, isFileCacheable bool) {
	opts := []grpc.DialOption{
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(utils.GetIntEnvironmentVariable("MAX_GRPC_MESSAGE_SIZE"))),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	ownerEdge := lookupResponse.OwnerEdge
	conn, err := grpc.Dial(ownerEdge.IpAddr, opts...)
	if err != nil {
		customErr := status.Error(codes.Code(client.ErrorCodes_STREAM_CLOSE_ERROR), fmt.Sprintf("[*DOWNLOAD_ERROR*] - Failed while trying to dial ownerEdge via gRPC.\r\nError: '%s'", err.Error()))
		clientRedirectionChannel.MessageChannel <- channels.Message{Body: []byte{}, Err: customErr}
		if isFileCacheable {
			cacheRedirectionChannel.MessageChannel <- channels.Message{Body: []byte{}, Err: customErr}
		}
		return
	}
	grpcClient := client.NewEdgeFileServiceClient(conn)
	context := context.Background()
	edgeDownloadStream, err := grpcClient.DownloadFromEdge(context, &client.FileDownloadRequest{TicketId: "", FileName: fileName})
	if err != nil {
		customErr := status.Error(codes.Code(client.ErrorCodes_STREAM_CLOSE_ERROR), fmt.Sprintf("[*DOWNLOAD_ERROR*] - Failed while triggering download from edge via gRPC.\r\nError: '%s'", err.Error()))
		clientRedirectionChannel.MessageChannel <- channels.Message{Body: []byte{}, Err: customErr}
		if isFileCacheable {
			cacheRedirectionChannel.MessageChannel <- channels.Message{Body: []byte{}, Err: customErr}
		}
		return
	}

	rcvAndRedirectChunks(clientRedirectionChannel, cacheRedirectionChannel, isFileCacheable, edgeDownloadStream)
	err = edgeDownloadStream.CloseSend()
	if err != nil {
		customErr := status.Error(codes.Code(client.ErrorCodes_STREAM_CLOSE_ERROR), fmt.Sprintf("[*ERROR*] - Impossibile chiudere downloadstream.\r\nError: '%s'", err.Error()))
		clientRedirectionChannel.MessageChannel <- channels.Message{Body: []byte{}, Err: customErr}
		if isFileCacheable {
			cacheRedirectionChannel.MessageChannel <- channels.Message{Body: []byte{}, Err: customErr}
		}
	}
}

// Invia il file al client direttamente dalla cache locale
func sendFromLocalCache(fileName string, clientDownloadStream client.FileService_DownloadServer) error {
	clientRedirectionChannel := channels.NewRedirectionChannel(utils.GetIntEnvironmentVariable("DOWNLOAD_CHANNEL_SIZE"))
	go readFromLocalCache(fileName, clientRedirectionChannel)
	redirectStreamToClient(clientRedirectionChannel, clientDownloadStream)
	return nil
}

func readFromLocalCache(fileName string, clientRedirectionChannel channels.RedirectionChannel) {
	utils.PrintEvent("CACHE", fmt.Sprintf("Il file '%s' è stato trovato nella cache locale", fileName))
	localFile, err := cache.GetCache().GetFileForReading(fileName)
	if err != nil {
		utils.PrintEvent("CACHE_ERROR", fmt.Sprintf("Impossibile aprire il file '%s'.", fileName))
		return
	}
	syscall.Flock(int(localFile.Fd()), syscall.F_RDLCK)
	defer syscall.Flock(int(localFile.Fd()), syscall.F_UNLCK)
	defer localFile.Close()

	chunkSize := utils.GetIntEnvironmentVariable("CHUNK_SIZE")
	buffer := make([]byte, chunkSize)

	defer close(clientRedirectionChannel.MessageChannel)
	for {
		bytesRead, err := localFile.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			clientRedirectionChannel.MessageChannel <- channels.Message{Body: []byte{}, Err: status.Error(codes.Code(client.ErrorCodes_FILE_READ_ERROR), fmt.Sprintf("[*ERROR*] - Failed during read operation.\r\nError: '%s'", err.Error()))}
		}
		clientRedirectionChannel.MessageChannel <- channels.Message{Body: buffer[:bytesRead], Err: nil}
	}
}
