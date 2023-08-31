package server

import (
	"crypto/sha256"
	"edge/peer"
	"edge/proto/client"
	"edge/proto/edge"
	"edge/utils"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"syscall"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"golang.org/x/net/context"

	"google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	status "google.golang.org/grpc/status"
)

type FileServiceServer struct {
	client.UnimplementedFileServiceServer
}

type DownloadStream struct {
	clientStream client.FileService_DownloadServer
	fileName     string
	ticketID     string
	fileChannel  chan []byte
}

type UploadStream struct {
	clientStream client.FileService_UploadServer
	fileName     string
	fileChannel  chan []byte
}

func (u *UploadStream) Read(p []byte) (n int, err error) {
	fileChunk, err := u.clientStream.Recv()
	if err != nil {
		errorHash := sha256.Sum256([]byte("[*ERROR*]"))
		u.fileChannel <- errorHash[:]
		return 0, fmt.Errorf("[*ERROR*] -> Stream redirection to S3 encountered some problems")
	}
	copy(p, fileChunk.Chunk)
	u.fileChannel <- fileChunk.Chunk

	return len(fileChunk.Chunk), nil
}

func (d *DownloadStream) WriteAt(p []byte, off int64) (n int, err error) {
	// Se il Concurrency del Downloader è impostato ad 1 non serve usare l'offset
	err = d.clientStream.Send(&client.FileChunk{Chunk: p})
	if err != nil {
		errorHash := sha256.Sum256([]byte("[*ERROR*]"))
		d.fileChannel <- errorHash[:]
		return 0, fmt.Errorf("[*ERROR*] -> Stream redirection to client encountered some problems")
	}
	log.Printf("[*] -> Loaded Chunk of size %d\n", len(p))
	chunkCopy := make([]byte, len(p)) // IMPORTANTE
	copy(chunkCopy, p)
	d.fileChannel <- chunkCopy
	return len(p), nil
}

func (s *FileServiceServer) Upload(uploadStream client.FileService_UploadServer) error {
	// Apri il file locale dove verranno scritti i chunks
	log.Println(uploadStream.Context().Value("FILE_NAME"))
	fileName := string(uploadStream.Context().Value("FILE_NAME").(string))
	ticketID := string(uploadStream.Context().Value("TICKET_ID").(string))
	md, thereIsMetadata := metadata.FromIncomingContext(uploadStream.Context())
	if !thereIsMetadata {
		return status.Error(codes.Code(client.ErrorCodes_INVALID_TICKET), "[*ERROR*] - Request No Ticket")
	}
	//md.Get()
	log.Println(md)
	return nil
	// TODO Vedere come gestire grandezza dei file --> Passiamo dimensione del file nel context

	isValidRequest := checkTicket(ticketID)
	if isValidRequest == -1 {
		return status.Error(codes.Code(client.ErrorCodes_INVALID_TICKET), "[*ERROR*] - Request with Invalid Ticket")
	}
	defer publishNewTicket(isValidRequest)

	// Salvo prima su file locale o su S3?? PRIMA S3
	fileChannel := make(chan []byte, utils.GetIntegerEnvironmentVariable("UPLOAD_CHANNEL_SIZE"))
	defer close(fileChannel)

	uploadStreamReader := UploadStream{clientStream: uploadStream, fileName: fileName, fileChannel: fileChannel}

	// Thread in attesa di ricevere il file durante l'invio ad S3 in modo da salvarlo localmente
	go writeChunksOnFile(fileChannel, fileName)
	sendToS3(fileName, uploadStreamReader)

	log.Printf("[*SUCCESS*] - File '%s' caricato con successo [TICKET_ID: %s]\r\n", fileName, ticketID)
	response := client.Response{TicketId: ticketID, Success: true}
	err := uploadStream.SendAndClose(&response)
	if err != nil {
		return status.Error(codes.Code(client.ErrorCodes_CHUNK_ERROR), "[*ERROR*] - Couldn't close clientstream")
	}

	return nil
}

// TODO impostare un timer dopo il quale assumo che il file non sia stato trovato --> limito l'attesa
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
		if err == nil { // ce l'ha ownerEdge --> Ricevi file come stream e invia chunk (+ salva in locale)
			log.Printf("[*NETWORK*] -> file '%s' found in edge %s", requestMessage.FileName, ownerEdge.PeerAddr)
			sendFromOtherEdge(ownerEdge, requestMessage, downloadStream)
		} else { // ce l'ha S3 --> Ricevi file come stream e invia chunk (+ salva in locale)
			log.Printf("[*S3*] -> looking for file '%s' in s3...", requestMessage.FileName)
			err = sendFromS3(requestMessage, downloadStream)
			if err != nil {
				return status.Error(codes.Code(client.ErrorCodes_FILE_NOT_FOUND_ERROR), "[*ERROR*] - Couldn't locate requested file in specified bucket")
			}
		}
	} else if err == nil { // ce l'ha l'edge corrente --> Leggi file e invia chunk
		log.Printf("[*CACHE*] -> File '%s' found in cache", requestMessage.FileName)
		return sendFromLocalFileStream(requestMessage, downloadStream)
	} else { //Got an error
		// TODO Aggiungere errore per path sbagliato

	}
	log.Println("[*SUCCESS*] -> Invio del file '" + requestMessage.FileName + "' Completata")
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
	fileChannel := make(chan []byte, utils.GetIntegerEnvironmentVariable("DOWNLOAD_CHANNEL_SIZE"))
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
		clientDownloadStream.Send(&client.FileChunk{Chunk: edgeChunk.Chunk})
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

func sendToS3(fileName string, uploadStreamReader UploadStream) error {

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Profile: "default",
	}))
	uploader := s3manager.NewUploader(sess, func(d *s3manager.Uploader) {
		d.PartSize = int64(utils.GetIntegerEnvironmentVariable("CHUNK_SIZE"))
		d.Concurrency = 1 //TODO Vedere se implementarlo in modo parallelo --> Serve numero d'ordine nel FileChunk
	})

	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(utils.GetEnvironmentVariable("S3_BUCKET_NAME")), // nome bucket
		Key:    &fileName,                                                  //percorso file da caricare
		Body:   &uploadStreamReader,
	})
	if err != nil {
		fmt.Println("Errore nell'upload:", err)
		return err
	}

	return nil
}

func sendFromS3(requestMessage *client.FileDownloadRequest, clientDownloadStream client.FileService_DownloadServer) error {
	// 1] Open connection to S3
	// 2] retrieve chunk by chunk (send to client + save in local)
	fileChannel := make(chan []byte, utils.GetIntegerEnvironmentVariable("DOWNLOAD_CHANNEL_SIZE"))
	downloadStreamWriter := DownloadStream{clientDownloadStream, requestMessage.FileName, requestMessage.TicketId, fileChannel}
	defer close(fileChannel)
	// sess := session.Must(session.NewSessionWithOptions(session.Options{
	// 	Profile:     "default",
	// 	Credentials: session.CredentialsProviderOptions{},
	// }))
	go writeChunksOnFile(fileChannel, requestMessage.FileName)
	// sess, err := session.NewSession(&aws.Config{
	// 	Region:      aws.String("us-east-1"),
	// 	Credentials: credentials.NewSharedCredentials("/home/.aws/credentials", ""),
	// })
	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String("us-east-1"),
		Credentials: credentials.NewSharedCredentials("/home/.aws/credentials", ""),
	}))
	// if err != nil {
	// 	errorHash := sha256.Sum256([]byte("[*ERROR*]"))
	// 	fileChannel <- errorHash[:]
	// 	return fmt.Errorf("[*ERROR*] Error Opening AWS Session")
	// }

	// TODO Capire perché fa come cazzo gli pare
	// Crea un downloader con la dimensione delle parti configurata
	downloader := s3manager.NewDownloader(
		sess,
		func(d *s3manager.Downloader) {
			d.PartSize = 10 * 1024 * 1024 // int64(utils.GetIntegerEnvironmentVariable("CHUNK_SIZE"))
			d.Concurrency = 1             //TODO Vedere se implementarlo in modo parallelo --> Serve numero d'ordine nel FileChunk
		},
	)
	// NOTA -> se ne sbatte il cazzo dei parametri :)
	log.Println(downloader.PartSize, downloader.Concurrency)

	// Esegui il download parallelo e scrivi i dati nello stream gRPC
	_, err := downloader.Download(
		&downloadStreamWriter,
		&s3.GetObjectInput{
			Bucket: aws.String(utils.GetEnvironmentVariable("S3_BUCKET_NAME")), //nome bucket
			Key:    aws.String(requestMessage.FileName),                        //percorso file da scaricare
		},
	)
	if err != nil {
		errorHash := sha256.Sum256([]byte("[*ERROR*]"))
		fileChannel <- errorHash[:]
		fmt.Println("Errore nel download:", err)
		return err
	}

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

func checkTicket(requestId string) int {
	for index, authRequestId := range authorizedTicketIDs.IDs {
		if requestId == authRequestId {
			return index
		}
	}
	return -1
}
