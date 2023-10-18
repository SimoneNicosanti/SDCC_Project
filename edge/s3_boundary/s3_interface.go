package s3_boundary

import (
	"edge/redirection_channel"
	"edge/utils"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const MB int64 = 1024 * 1024

func SendToS3(fileName string, redirectionChannel redirection_channel.RedirectionChannel) {
	uploadStreamReader := UploadStream{RedirectionChannel: redirectionChannel, ResidualChunk: make([]byte, 0)}
	defer close(redirectionChannel.ReturnChannel)

	sess := getSession()
	uploader := s3manager.NewUploader(sess, func(d *s3manager.Uploader) {
		d.PartSize = getS3ChunkSize("S3_UPLOAD_CHUNK_SIZE")
		d.Concurrency = 1
	})
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(utils.GetEnvironmentVariable("S3_BUCKET_NAME")), // nome bucket
		Key:    aws.String(fileName),                                       //percorso file da caricare
		Body:   &uploadStreamReader,
	})
	redirectionChannel.ReturnChannel <- err

}

func SendFromS3(fileName string, clientRedirectionChannel redirection_channel.RedirectionChannel, cacheRedirectionChannel redirection_channel.RedirectionChannel, isFileCachable bool) error {
	// 1] Open connection to S3
	// 2] retrieve chunk by chunk (send to client + save in local)
	downloadStreamWriter := DownloadStream{
		ClientChannel:   clientRedirectionChannel,
		CacheChannel:    cacheRedirectionChannel,
		IsFileCacheable: isFileCachable,
		DownloadBuffer:  []byte{},
		FlushSize:       getS3ChunkSize("S3_DOWNLOAD_CHUNK_SIZE"),
	}
	defer close(cacheRedirectionChannel.MessageChannel)
	defer close(clientRedirectionChannel.MessageChannel)

	sess := getSession()

	inputObject := s3.GetObjectInput{
		Bucket: aws.String(utils.GetEnvironmentVariable("S3_BUCKET_NAME")), //nome bucket
		Key:    aws.String(fileName),                                       //percorso file da scaricare
	}

	// Crea un downloader con la dimensione delle parti configurata
	// RequestOptions s3.request.Option
	downloader := s3manager.NewDownloader(
		sess,
		func(d *s3manager.Downloader) {
			d.PartSize = getS3ChunkSize("S3_DOWNLOAD_CHUNK_SIZE") // //TODO capire perché non rispetta il parametro
			d.Concurrency = 1
		},
	)

	// Esegui il download e scrivi i dati nello stream gRPC
	_, err := downloader.Download(
		&downloadStreamWriter,
		&inputObject,
	)
	if err != nil {
		fmt.Println("ERRORE_0")
		customErr := fmt.Errorf("Errore nella download del file '%s'\r\nL'errore restituito è: '%s'", fileName, err.Error())
		clientRedirectionChannel.MessageChannel <- redirection_channel.Message{Body: []byte{}, Err: customErr}
		if isFileCachable {
			cacheRedirectionChannel.MessageChannel <- redirection_channel.Message{Body: []byte{}, Err: customErr}
		}
		return err
	}

	_, err = (&downloadStreamWriter).Flush()
	if err != nil {
		fmt.Println("ERRORE_1")
		customErr := fmt.Errorf("Errore nella download del file '%s'\r\nL'errore restituito è: '%s'", fileName, err.Error())
		clientRedirectionChannel.MessageChannel <- redirection_channel.Message{Body: []byte{}, Err: customErr}
		if isFileCachable {
			cacheRedirectionChannel.MessageChannel <- redirection_channel.Message{Body: []byte{}, Err: customErr}
		}
		return err
	}

	return nil
}

func DeleteFromS3(fileName string) error {
	sess := getSession()
	svc := s3.New(sess)

	_, err := getHeadObject(fileName)
	if err != nil {
		utils.PrintEvent("S3_DELETE_ERR", "Impossibile verificare esistenza del file")
		aerr := err.(awserr.Error)
		if strings.Compare(aerr.Code(), "NotFound") == 0 {
			return fmt.Errorf("NotFound")
		}
		return err
	}

	_, err = svc.DeleteObject(
		&s3.DeleteObjectInput{
			Bucket: aws.String(utils.GetEnvironmentVariable("S3_BUCKET_NAME")),
			Key:    aws.String(fileName),
		},
	)
	if err != nil {
		utils.PrintEvent("S3_DELETE_ERR", fmt.Sprintf("Impossibile eliminare il file '%s'. Error: '%s", fileName, err.Error()))
		return err
	}

	err = svc.WaitUntilObjectNotExists(&s3.HeadObjectInput{
		Bucket: aws.String(utils.GetEnvironmentVariable("S3_BUCKET_NAME")),
		Key:    aws.String(fileName),
	})
	if err != nil {
		utils.PrintEvent("S3_DELETE_WAIT_ERR", fmt.Sprintf("Impossibile verificare l'eliminazione del file '%s'. Error: '%s", fileName, err.Error()))
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

func GetFileSize(fileName string) (int64, error) {
	headObjOutput, err := getHeadObject(fileName)
	if err != nil {
		return -1, err
	}
	fileSize := headObjOutput.ContentLength
	return *fileSize, nil
}

func getHeadObject(fileName string) (*s3.HeadObjectOutput, error) {
	sess := getSession()
	svc := s3.New(sess)

	headObjOutput, err := svc.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(utils.GetEnvironmentVariable("S3_BUCKET_NAME")),
		Key:    aws.String(fileName),
	})
	if err != nil {
		return nil, err
	}
	return headObjOutput, nil
}

func getS3ChunkSize(chunkTypeStr string) int64 {
	envVariable := utils.GetInt64EnvironmentVariable(chunkTypeStr)
	if envVariable < 5 {
		return 5 * MB
	}
	return envVariable * MB
}
