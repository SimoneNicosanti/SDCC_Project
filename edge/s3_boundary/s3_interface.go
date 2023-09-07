package s3_boundary

import (
	"crypto/sha256"
	"edge/proto/client"
	"edge/utils"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func SendToS3(fileName string, uploadStreamReader UploadStream) error {

	sess := getSession()
	uploader := s3manager.NewUploader(sess, func(d *s3manager.Uploader) {
		d.PartSize = getS3ChunkSize("S3_UPLOAD_CHUNK_SIZE")
		d.Concurrency = 1 //TODO Vedere se implementarlo in modo parallelo --> Serve numero d'ordine nel FileChunk
	})
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(utils.GetEnvironmentVariable("S3_BUCKET_NAME")), // nome bucket
		Key:    aws.String(fileName),                                       //percorso file da caricare
		Body:   &uploadStreamReader,
	})
	if err != nil {
		errorHash := sha256.Sum256([]byte("[*ERROR*]"))
		uploadStreamReader.FileChannel <- errorHash[:]
		fmt.Println("Errore nell'upload:", err)
		return err
	}

	return nil
}

func SendFromS3(requestMessage *client.FileDownloadRequest, downloadStreamWriter DownloadStream) error {
	// 1] Open connection to S3
	// 2] retrieve chunk by chunk (send to client + save in local)
	sess := getSession()
	fileSize, err := getFileSize(sess, requestMessage.FileName)
	if err != nil {
		return err
	}

	if fileSize > int64(utils.GetIntEnvironmentVariable("CACHE_FILE_MAX_SIZE")) {
		//TODO Rispondi solo al file ma non metti in Cache
	} else {
		//TODO Metti in cache
	}
	// Crea un downloader con la dimensione delle parti configurata
	downloader := s3manager.NewDownloader(
		sess,
		func(d *s3manager.Downloader) {
			d.PartSize = getS3ChunkSize("S3_DOWNLOAD_CHUNK_SIZE") //TODO capire perché non rispetta il parametro
			d.Concurrency = 1                                     //TODO Vedere se implementarlo in modo parallelo --> Serve numero d'ordine nel FileChunk
		},
	)
	log.Println(downloader.PartSize, downloader.Concurrency)

	// Esegui il download e scrivi i dati nello stream gRPC
	_, err = downloader.Download(
		&downloadStreamWriter,
		&s3.GetObjectInput{
			Bucket: aws.String(utils.GetEnvironmentVariable("S3_BUCKET_NAME")), //nome bucket
			Key:    aws.String(requestMessage.FileName),                        //percorso file da scaricare
		},
	)
	if err != nil {
		fmt.Println("Errore nel download:", err)
		return err
	}

	return nil
}

func getSession() *session.Session {
	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String("us-east-1"),
		Credentials: credentials.NewSharedCredentials("/aws/credentials", ""),
	}))
	return sess
}

func getFileSize(sess *session.Session, fileName string) (int64, error) {
	svc := s3.New(sess)
	headObjOutput, err := svc.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(utils.GetEnvironmentVariable("S3_BUCKET_NAME")),
		Key:    aws.String(fileName),
	})
	if err != nil {
		return -1, err
	}
	fileSize := headObjOutput.ContentLength
	return *fileSize, nil
}

func getS3ChunkSize(chunkTypeStr string) int64 {
	envVariable := utils.GetIntEnvironmentVariable(chunkTypeStr)
	if int64(envVariable) < 5 {
		return 5 * 1024 * 1024
	}
	return int64(envVariable) * 1024 * 1024
}
