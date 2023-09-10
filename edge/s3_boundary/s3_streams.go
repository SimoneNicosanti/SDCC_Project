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
		utils.PrintEvent("UPLOAD_RCV", "Letto da Residuo")
		fileChunk = uploadStream.ResidualChunk
	} else {
		select {
		case chunk, _ := <-uploadStream.RedirectionChannel.ChunkChannel:
			utils.PrintEvent("UPLOAD_RCV", "Letto da Chunk")
			if len(chunk) == 0 {
				// Potrebbe essere stato chiuso a causa di un errore
				utils.PrintEvent("UPLOAD_CLOSE", "Canale Chiuso")
				err := <-uploadStream.RedirectionChannel.ErrorChannel
				if err != nil {
					return 0, err
				}
				return 0, io.EOF
			}
			fileChunk = chunk
		case err := <-uploadStream.RedirectionChannel.ErrorChannel:
			if err != nil {
				utils.PrintEvent("UPLOAD_ERR", "Ricevuto Errore")
				return 0, err
			}
		}
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
	// Se il Concurrency del Downloader è impostato ad 1 non serve usare l'offset
	// TODO Vedere se bisogna aggiungere la Select
	clientCopy := make([]byte, len(source))
	copy(clientCopy, source)
	downloadStream.ClientChannel.ChunkChannel <- clientCopy

	if downloadStream.IsFileCacheable {
		//log.Printf("[*] -> Loaded Chunk of size %d\n", len(p))
		cacheCopy := make([]byte, len(source)) // IMPORTANTE
		copy(cacheCopy, source)
		downloadStream.CacheChannel.ChunkChannel <- cacheCopy
	}
	return len(source), nil
}
