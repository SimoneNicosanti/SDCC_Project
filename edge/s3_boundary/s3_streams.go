package s3_boundary

import (
	"edge/channels"
	"edge/utils"
	"io"
)

type DownloadStream struct {
	ClientChannel   channels.RedirectionChannel
	CacheChannel    channels.RedirectionChannel
	IsFileCacheable bool
}

type UploadStream struct {
	ResidualChunk      []byte
	RedirectionChannel channels.RedirectionChannel
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
	//TODO aggiungere controllo sui canali di errore dei riceventi
	//TODO (se c'è errore -> stop se è sul canale principale altrimenti smetti di inviare sulla cache e bassta)

	// Se il Concurrency del Downloader è impostato ad 1 non serve usare l'offset
	clientCopy := make([]byte, len(source))
	copy(clientCopy, source)
	downloadStream.ClientChannel.MessageChannel <- channels.Message{Body: clientCopy, Err: nil}

	if downloadStream.IsFileCacheable {
		//log.Printf("[*] -> Loaded Chunk of size %d\n", len(p))
		cacheCopy := make([]byte, len(source))
		copy(cacheCopy, source)
		downloadStream.CacheChannel.MessageChannel <- channels.Message{Body: cacheCopy, Err: nil}
	}
	return len(source), nil
}
