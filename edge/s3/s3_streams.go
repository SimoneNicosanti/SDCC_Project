package s3

import (
	"crypto/sha256"
	"edge/proto/client"
	"fmt"
	"io"
	"log"
)

type DownloadStream struct {
	ClientStream client.FileService_DownloadServer
	FileName     string
	TicketID     string
	FileChannel  chan []byte
}

type UploadStream struct {
	ClientStream  client.FileService_UploadServer
	FileName      string
	ResidualChunk []byte
	FileChannel   chan []byte
}

func (u *UploadStream) Read(p []byte) (n int, err error) {

	var fileChunk []byte

	if len(u.ResidualChunk) > 0 {
		// Parte del chunk precedente deve essere consumata
		log.Println("[*UPLOAD*] -> Letto da Residuo")
		fileChunk = u.ResidualChunk
	} else {
		log.Println("[*UPLOAD*] -> Letto da Canale")
		chunkMessage, err := u.ClientStream.Recv()
		if err == io.EOF {
			return 0, io.EOF
		}
		if err != nil {
			errorHash := sha256.Sum256([]byte("[*ERROR*]"))
			u.FileChannel <- errorHash[:]
			return 0, fmt.Errorf("[*ERROR*] -> Message Receive From gRPC encountered some problems")
		}

		// Send chunk to local Write
		chunkCopy := make([]byte, len(chunkMessage.Chunk))
		copy(chunkCopy, chunkMessage.Chunk)
		u.FileChannel <- chunkCopy

		fileChunk = chunkMessage.Chunk
	}

	pLen := len(p)
	chunkLen := len(fileChunk)
	if pLen > chunkLen {
		copy(p, fileChunk)
		p = p[0:chunkLen]
		u.ResidualChunk = u.ResidualChunk[0:0]
	} else {
		copy(p, fileChunk[0:pLen])
		u.ResidualChunk = fileChunk[pLen:]
	}

	return len(p), nil
}

func (d *DownloadStream) WriteAt(p []byte, off int64) (n int, err error) {
	// Se il Concurrency del Downloader è impostato ad 1 non serve usare l'offset
	err = d.ClientStream.Send(&client.FileChunk{Chunk: p})
	if err != nil {
		errorHash := sha256.Sum256([]byte("[*ERROR*]"))
		d.FileChannel <- errorHash[:]
		return 0, fmt.Errorf("[*ERROR*] -> Stream redirection to client encountered some problems")
	}
	log.Printf("[*] -> Loaded Chunk of size %d\n", len(p))
	chunkCopy := make([]byte, len(p)) // IMPORTANTE
	copy(chunkCopy, p)
	d.FileChannel <- chunkCopy
	return len(p), nil
}
