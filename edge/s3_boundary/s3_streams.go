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
	DownloadBuffer  []byte
	FlushSize       int64
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

func (downloadStream *DownloadStream) WriteAt(source []byte, off int64) (sent int, err error) {
	select {
	case err := <-downloadStream.ClientChannel.ReturnChannel:
		return -1, err
	case <-downloadStream.CacheChannel.ReturnChannel:
		downloadStream.IsFileCacheable = false
	default:
		break
	}
	// since we have no guarantee on 1:1 correspondence on what is passed to WriteAt we buffer the source until the desired size
	// but we still have problems with huge download times...
	if len(downloadStream.DownloadBuffer)+len(source) < int(downloadStream.FlushSize) {
		fmt.Printf("Buffered %d bytes\n", len(source))
		downloadStream.DownloadBuffer = append(downloadStream.DownloadBuffer, source...)
		return len(source), nil
	}
	_, err = downloadStream.writeAndRedirect(downloadStream.DownloadBuffer[0:len(downloadStream.DownloadBuffer)])
	fmt.Printf("FLUSHED %d bytes\n", len(downloadStream.DownloadBuffer))
	downloadStream.DownloadBuffer = downloadStream.DownloadBuffer[:]
	return 0, err
}

func (downloadStream *DownloadStream) Flush() (int, error) {
	if len(downloadStream.DownloadBuffer) > 0 {
		return downloadStream.writeAndRedirect(downloadStream.DownloadBuffer[0:len(downloadStream.DownloadBuffer)])
	}
	return 0, nil
}

func (downloadStream *DownloadStream) writeAndRedirect(source []byte) (redirected int, err error) {
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
