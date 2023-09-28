package s3_boundary

import (
	redirectionchannel "edge/redirection_channel"
	"edge/utils"
	"fmt"
	"io"
)

type DownloadStream struct {
	ClientChannel   redirectionchannel.RedirectionChannel
	CacheChannel    redirectionchannel.RedirectionChannel
	IsFileCacheable bool
}

type UploadStream struct {
	ResidualChunk      []byte
	RedirectionChannel redirectionchannel.RedirectionChannel
}

func (uploadStream *UploadStream) Read(dest []byte) (bytesInDest int, err error) {

	var fileChunk []byte

	if len(uploadStream.ResidualChunk) > 0 {
		// Parte del chunk precedente deve essere consumata
		//utils.PrintEvent("UPLOAD_RCV", "Letto da Residuo")
		fileChunk = uploadStream.ResidualChunk
	} else {
		message, isOpen := <-uploadStream.RedirectionChannel.MessageChannel
		if message.Err != nil {
			utils.PrintEvent("UPLOAD_ERR", "Ricevuto Errore")
			return 0, message.Err
		}
		if !isOpen {
			return 0, io.EOF
		}
		fileChunk = message.Body
	}

	destLen := len(dest)
	chunkLen := len(fileChunk)
	if destLen > chunkLen {
		copy(dest, fileChunk)
		dest = dest[0:chunkLen]
		uploadStream.ResidualChunk = uploadStream.ResidualChunk[0:0]
	} else {
		copy(dest, fileChunk[0:destLen])
		uploadStream.ResidualChunk = fileChunk[destLen:]
	}

	return len(dest), nil
}

func (downloadStream *DownloadStream) WriteAt(source []byte, off int64) (bytesSent int, err error) {
	select {
	case err := <-downloadStream.ClientChannel.ReturnChannel:
		return 0, err
	case <-downloadStream.CacheChannel.ReturnChannel:
		downloadStream.IsFileCacheable = false
	default:
		break
	}
	return downloadStream.writeAndRedirect(source)
}

func (downloadStream *DownloadStream) writeAndRedirect(source []byte) (int, error) {
	clientCopy := make([]byte, len(source))
	copy(clientCopy, source)
	downloadStream.ClientChannel.MessageChannel <- redirectionchannel.Message{Body: clientCopy, Err: nil}

	if downloadStream.IsFileCacheable {

		cacheCopy := make([]byte, len(source))
		copy(cacheCopy, source)
		downloadStream.CacheChannel.MessageChannel <- redirectionchannel.Message{Body: cacheCopy, Err: nil}
	}
	fmt.Printf("INVIO CHUNK DI LUNGHEZZA %d\n", len(source))
	return len(source), nil
}
