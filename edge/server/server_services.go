package server

import (
	"context"
	"edge/cache"
	"edge/peer"
	"edge/proto/file_transfer"
	"edge/redirection_channel"
	"edge/s3_boundary"
	"edge/utils"
	"fmt"
	"io"
	"strconv"
	"syscall"
	"time"

	"google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	status "google.golang.org/grpc/status"
)

func (s *FileServiceServer) Delete() error {
	s.incrementWorkload()
	returnValue := doDelete()
	s.decrementWorkload()
	return returnValue
}

func doDelete() error {
	//TODO
	return nil
}

func (s *FileServiceServer) Upload(uploadStream file_transfer.FileService_UploadServer) error {
	s.incrementWorkload()
	returnValue := doUpload(uploadStream)
	s.decrementWorkload()
	return returnValue
}

func doUpload(uploadStream file_transfer.FileService_UploadServer) error {
	// Apri il file locale dove verranno scritti i chunks
	defer notifyJobEnd()

	requestID, fileName, fileSize, err := retrieveMetadata(uploadStream)
	if err != nil {
		return err
	}
	utils.PrintEvent("CLIENT_REQUEST_RECEIVED", fmt.Sprintf("Ricevuta richiesta di upload per file '%s'\r\nRequest ID: '%s'", fileName, requestID))

	// Salvo prima su S3 e poi su file locale
	s3RedirectionChannel := redirection_channel.NewRedirectionChannel(utils.GetIntEnvironmentVariable("UPLOAD_CHANNEL_SIZE"))
	cacheRedirectionChannel := redirection_channel.NewRedirectionChannel(utils.GetIntEnvironmentVariable("UPLOAD_CHANNEL_SIZE"))

	// Ridirezione del flusso sulla cache
	isFileCacheable := cache.IsFileCacheable(fileSize)
	if isFileCacheable {
		// Thread in attesa di ricevere il file durante l'invio ad S3 in modo da salvarlo localmente
		go cache.GetCache().InsertFileInCache(cacheRedirectionChannel, fileName, fileSize)
	}

	// Ridirezione su S3
	go s3_boundary.SendToS3(fileName, s3RedirectionChannel)

	err = rcvAndRedirectChunks(s3RedirectionChannel, cacheRedirectionChannel, isFileCacheable, uploadStream)

	if err != nil {
		utils.PrintEvent("UPLOAD_ERROR", fmt.Sprintf("Errore nel caricare il file '%s'\r\nRequest ID: '%s'", fileName, requestID))
		return status.Error(codes.Code(file_transfer.ErrorCodes_S3_ERROR), fmt.Sprintf("[*ERROR*] - File Upload to S3 encountered some error.\r\nError: '%s'", err.Error()))
	}
	utils.PrintEvent("UPLOAD_SUCCESS", fmt.Sprintf("File '%s' caricato con successo\r\nRequest ID: '%s'", fileName, requestID))
	response := file_transfer.FileResponse{RequestId: requestID, Success: true}
	err = uploadStream.SendAndClose(&response)
	if err != nil {
		return status.Error(codes.Code(file_transfer.ErrorCodes_CHUNK_ERROR), fmt.Sprintf("[*ERROR*] - Impossibile chiudere il clientstream.\r\nError: '%s'", err.Error()))
	}

	return nil
}

func retrieveMetadata(uploadStream file_transfer.FileService_UploadServer) (requestID string, file_name string, file_size int64, err error) {
	md, thereIsMetadata := metadata.FromIncomingContext(uploadStream.Context())
	if !thereIsMetadata {
		return "", "", 0, status.Error(codes.Code(file_transfer.ErrorCodes_INVALID_METADATA), "[*NO_METADATA*] - No metadata found")
	}
	requestID = md.Get("request_id")[0]
	file_name = md.Get("file_name")[0]
	file_size, err = strconv.ParseInt(md.Get("file_size")[0], 10, 64)
	if err != nil {
		return "", "", 0, status.Error(codes.Code(file_transfer.ErrorCodes_INVALID_METADATA), fmt.Sprintf("[*CAST_ERROR*] - Impossibile effettuare il cast della size : '%s'", err.Error()))
	}
	return requestID, file_name, file_size, nil
}

// Permette al client di effettuare una richiesta di get con successo
func (s *FileServiceServer) Download(requestMessage *file_transfer.FileDownloadRequest, downloadStream file_transfer.FileService_DownloadServer) error {
	s.incrementWorkload()
	returnValue := doDownload(requestMessage, downloadStream)
	s.decrementWorkload()
	return returnValue
}

func doDownload(requestMessage *file_transfer.FileDownloadRequest, downloadStream file_transfer.FileService_DownloadServer) error {
	defer notifyJobEnd()

	utils.PrintEvent("CLIENT_REQUEST_RECEIVED", fmt.Sprintf("Ricevuta richiesta di download per file '%s'\r\nRequestID: '%s'", requestMessage.FileName, requestMessage.RequestId))

	if cache.GetCache().IsFileInCache(requestMessage.FileName) {
		// ce l'ha l'edge corrente --> Leggi file e invia chunk
		return sendFromLocalCache(requestMessage.FileName, downloadStream)
	} else {
		// Thread in attesa di ricevere il file durante l'invio ad S3 in modo da salvarlo localmente
		lookupServer, err := peer.CreateLookupServer()

		if err != nil {
			utils.PrintEvent("LOOKUP_SERVER_ERROR", "Impossibile creare il lookup server")
			return err
		}
		callbackChannel := lookupFileInNetwork(lookupServer, requestMessage)

		// potrebbe averlo qualche edge --> attendi risposte alla lookup
		var askS3 bool = !tryToSendFromOtherEdge(callbackChannel, requestMessage, downloadStream)
		//lookupServer.CloseServer()

		if askS3 { // nessun edge contattato ha il file --> Contatta S3: Ricevi file come stream e invia chunk (+ salva in locale)
			err := redirectFromS3(requestMessage.FileName, downloadStream)
			if err != nil {
				utils.PrintEvent("S3_DOWNLOAD_ERROR", err.Error())
				return status.Error(codes.Code(file_transfer.ErrorCodes_FILE_NOT_FOUND_ERROR), fmt.Sprintf("[*ERROR*] -> Impossibile recuperare il file '%s' dal bucket specificato.\r\nError: '%s'", requestMessage.FileName, err.Error()))
			}
		}
	}
	utils.PrintEvent("DOWNLOAD_SUCCESS", fmt.Sprintf("Invio del file '%s' Completata", requestMessage.FileName))
	return nil
}

// Attendi risposte sul callback channel e, se ricevute, prova a richiedere il file all'edge. Ritorna true se siamo riusciti ad inviare il file da un altro edge; false altrimenti.
func tryToSendFromOtherEdge(callbackChannel chan *peer.FileLookupResponse, requestMessage *file_transfer.FileDownloadRequest, downloadStream file_transfer.FileService_DownloadServer) bool {
	timer := time.After(time.Second * time.Duration(utils.GetInt64EnvironmentVariable("MAX_WAITING_TIME_FOR_EDGE")))

	for {
		select {
		case lookupReponse := <-callbackChannel: // ce l'ha ownerEdge --> Contatta l'edge che ti ha risposto, ricevi file come stream e invia chunk (+ salva in locale)
			err := sendFromOtherEdge(*lookupReponse, requestMessage.FileName, downloadStream)
			if err == nil {
				// Se il download del file da un altro edge è andato a buon fine, non contattare S3 e smetti di aspettare ulteriori risposte
				utils.PrintEvent("OTHEREDGE_DOWNLOAD_SUCCESS", fmt.Sprintf("Inviato il file '%s' da un altro edge.", requestMessage.FileName))
				return true
			} else {
				utils.PrintEvent("OTHEREDGE_DOWNLOAD_ERROR", fmt.Sprintf("Impossibile recuperare file '%s' da altro edge... Ripiego su S3\r\nError: '%s'", requestMessage.FileName, err.Error()))
			}
		case <-timer:
			utils.PrintEvent("TIMEOUT_ERROR", fmt.Sprintf("Timeout nell'attesa per la ricerca del file '%s'. Nessuno ha risposto, ripiego su S3", requestMessage.FileName))
			return false
		}
	}
}

func redirectFromS3(fileName string, downloadStream file_transfer.FileService_DownloadServer) error {
	utils.PrintEvent("S3_LOOKUP", fmt.Sprintf("Cercando il file '%s' in s3...", fileName))
	var isFileCachable bool = false
	fileSize, err := s3_boundary.GetFileSize(fileName)
	if err != nil {
		utils.PrintEvent("S3_ERROR", fmt.Sprintf("Impossibile ricavare dimensione del file '%s' da S3", fileName))
	} else {
		isFileCachable = cache.IsFileCacheable(fileSize)
	}

	clientRedirectionChannel := redirection_channel.NewRedirectionChannel(utils.GetIntEnvironmentVariable("DOWNLOAD_CHANNEL_SIZE"))
	cacheRedirectionChannel := redirection_channel.NewRedirectionChannel(utils.GetIntEnvironmentVariable("DOWNLOAD_CHANNEL_SIZE"))

	if isFileCachable {
		go cache.GetCache().InsertFileInCache(cacheRedirectionChannel, fileName, fileSize)
	}

	go s3_boundary.SendFromS3(fileName, clientRedirectionChannel, cacheRedirectionChannel, isFileCachable)

	err = redirectStreamToClient(clientRedirectionChannel, downloadStream)

	return err
}

// Invia richieste di lookup ad alcuni dei tuoi vicini. Ritorna un channel dal quale possono essere lette le risposte alla file lookup.
func lookupFileInNetwork(lookupServer *peer.LookupServer, requestMessage *file_transfer.FileDownloadRequest) chan *peer.FileLookupResponse {

	peer.NeighboursFileLookupWithoutFileRequest(
		requestMessage.FileName,
		utils.GetIntEnvironmentVariable("REQUEST_TTL"),
		requestMessage.RequestId,
		peer.SelfPeer.PeerAddr,
		peer.SelfPeer.PeerAddr,
		lookupServer.UdpAddr)

	callbackChannel := make(chan *peer.FileLookupResponse, 10)
	go lookupServer.ReadFromServer(callbackChannel)

	return callbackChannel
}

func sendFromOtherEdge(lookupResponse peer.FileLookupResponse, fileName string, clientDownloadStream file_transfer.FileService_DownloadServer) error {
	utils.PrintEvent("FILE_IN_NETWORK", fmt.Sprintf("Il file '%s' è stato trovato nell'edge %s", fileName, lookupResponse.OwnerEdge.IpAddr))
	// 1] Open gRPC connection to ownerEdge
	// 2] retrieve chunk by chunk (send to client + save in local)
	clientRedirectionChannel := redirection_channel.NewRedirectionChannel(utils.GetIntEnvironmentVariable("DOWNLOAD_CHANNEL_SIZE"))
	cacheRedirectionChannel := redirection_channel.NewRedirectionChannel(utils.GetIntEnvironmentVariable("DOWNLOAD_CHANNEL_SIZE"))

	isFileCachable := cache.IsFileCacheable(lookupResponse.FileSize)
	if isFileCachable {
		go cache.GetCache().InsertFileInCache(cacheRedirectionChannel, fileName, lookupResponse.FileSize)
	}
	// Imposta la nuova dimensione massima
	go downloadFromOtherEdge(lookupResponse, fileName, cacheRedirectionChannel, clientRedirectionChannel, isFileCachable)

	return redirectStreamToClient(clientRedirectionChannel, clientDownloadStream)
}

func downloadFromOtherEdge(lookupResponse peer.FileLookupResponse, fileName string, cacheRedirectionChannel redirection_channel.RedirectionChannel, clientRedirectionChannel redirection_channel.RedirectionChannel, isFileCacheable bool) {
	opts := []grpc.DialOption{
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(utils.GetIntEnvironmentVariable("MAX_GRPC_MESSAGE_SIZE"))),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	ownerEdge := lookupResponse.OwnerEdge
	conn, err := grpc.Dial(ownerEdge.IpAddr, opts...)
	if err != nil {
		customErr := status.Error(codes.Code(file_transfer.ErrorCodes_STREAM_CLOSE_ERROR), fmt.Sprintf("[*DOWNLOAD_ERROR*] - Failed while trying to dial ownerEdge via gRPC.\r\nError: '%s'", err.Error()))
		clientRedirectionChannel.MessageChannel <- redirection_channel.Message{Body: []byte{}, Err: customErr}
		if isFileCacheable {
			cacheRedirectionChannel.MessageChannel <- redirection_channel.Message{Body: []byte{}, Err: customErr}
		}
		return
	}
	grpcClient := file_transfer.NewEdgeFileServiceClient(conn)
	context := context.Background()
	edgeDownloadStream, err := grpcClient.DownloadFromEdge(context, &file_transfer.FileDownloadRequest{RequestId: "", FileName: fileName})
	if err != nil {
		customErr := status.Error(codes.Code(file_transfer.ErrorCodes_STREAM_CLOSE_ERROR), fmt.Sprintf("[*DOWNLOAD_ERROR*] - Failed while triggering download from edge via gRPC.\r\nError: '%s'", err.Error()))
		clientRedirectionChannel.MessageChannel <- redirection_channel.Message{Body: []byte{}, Err: customErr}
		if isFileCacheable {
			cacheRedirectionChannel.MessageChannel <- redirection_channel.Message{Body: []byte{}, Err: customErr}
		}
		return
	}

	rcvAndRedirectChunks(clientRedirectionChannel, cacheRedirectionChannel, isFileCacheable, edgeDownloadStream)
	err = edgeDownloadStream.CloseSend()
	if err != nil {
		customErr := status.Error(codes.Code(file_transfer.ErrorCodes_STREAM_CLOSE_ERROR), fmt.Sprintf("[*ERROR*] - Impossibile chiudere downloadstream.\r\nError: '%s'", err.Error()))
		clientRedirectionChannel.MessageChannel <- redirection_channel.Message{Body: []byte{}, Err: customErr}
		if isFileCacheable {
			cacheRedirectionChannel.MessageChannel <- redirection_channel.Message{Body: []byte{}, Err: customErr}
		}
	}
}

// Invia il file al client direttamente dalla cache locale
func sendFromLocalCache(fileName string, clientDownloadStream file_transfer.FileService_DownloadServer) error {
	clientRedirectionChannel := redirection_channel.NewRedirectionChannel(utils.GetIntEnvironmentVariable("DOWNLOAD_CHANNEL_SIZE"))
	go readFromLocalCache(fileName, clientRedirectionChannel)
	redirectStreamToClient(clientRedirectionChannel, clientDownloadStream)
	return nil
}

// Recupera il file dalla cache locale e ne legge i chunk inviandoli sul canale dato in input
func readFromLocalCache(fileName string, clientRedirectionChannel redirection_channel.RedirectionChannel) {
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
			clientRedirectionChannel.MessageChannel <- redirection_channel.Message{Body: []byte{}, Err: status.Error(codes.Code(file_transfer.ErrorCodes_FILE_READ_ERROR), fmt.Sprintf("[*ERROR*] - Failed during read operation.\r\nError: '%s'", err.Error()))}
		}
		clientRedirectionChannel.MessageChannel <- redirection_channel.Message{Body: buffer[:bytesRead], Err: nil}
	}
}
