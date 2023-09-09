package s3_boundary

import (
	"crypto/sha256"
	"edge/channels"
	"edge/proto/client"
	"edge/utils"
	"fmt"
	"io"
)

type DownloadStream struct {
	ClientStream   client.FileService_DownloadServer
	FileName       string
	TicketID       string
	FileChannel    chan []byte
	WriteOnChannel bool
}

type UploadStream struct {
	ResidualChunk      []byte
	RedirectionChannel channels.RedirectionChannel
}

func (u *UploadStream) Read(dest []byte) (bytesInDest int, err error) {

	var fileChunk []byte

	if len(u.ResidualChunk) > 0 {
		// Parte del chunk precedente deve essere consumata
		utils.PrintEvent("UPLOAD_RCV", "Letto da Residuo")
		fileChunk = u.ResidualChunk
	} else {
		select {
		case chunk, _ := <-u.RedirectionChannel.ChunkChannel:
			utils.PrintEvent("UPLOAD_RCV", "Letto da Chunk")
			if len(chunk) == 0 {
				// Potrebbe essere stato chiuso a causa di un errore
				utils.PrintEvent("UPLOAD_CLOSE", "Canale Chiuso")
				err := <-u.RedirectionChannel.ErrorChannel
				if err != nil {
					return 0, err
				}
				return 0, io.EOF
			}
			fileChunk = chunk
		case err := <-u.RedirectionChannel.ErrorChannel:
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
		u.ResidualChunk = u.ResidualChunk[0:0]
	} else {
		copy(dest, fileChunk[0:destLen])
		u.ResidualChunk = fileChunk[destLen:]
	}

	return len(dest), nil
}

func (d *DownloadStream) WriteAt(p []byte, off int64) (n int, err error) {
	// Se il Concurrency del Downloader è impostato ad 1 non serve usare l'offset
	err = d.ClientStream.Send(&client.FileChunk{Chunk: p})
	if err != nil {
		errorHash := sha256.Sum256([]byte("[*ERROR*]"))
		d.FileChannel <- errorHash[:]
		return 0, fmt.Errorf("[*ERROR*] -> Stream redirection to client encountered some problems")
	}

	if d.WriteOnChannel {
		//log.Printf("[*] -> Loaded Chunk of size %d\n", len(p))
		chunkCopy := make([]byte, len(p)) // IMPORTANTE
		copy(chunkCopy, p)
		d.FileChannel <- chunkCopy
	}
	return len(p), nil
}
