package server

import (
	"edge/channels"
	"edge/proto/client"
	"fmt"
	"io"
)

func rcvAndRedirectChunks(s3RedirectionChannel channels.RedirectionChannel, cacheRedirectionChannel channels.RedirectionChannel, isFileCacheable bool, uploadStream client.FileService_UploadServer) error {
	var endLoop bool = false
	var s3Error error = nil
	var grpcError error = nil

	for !endLoop {
		select {
		case err := <-s3RedirectionChannel.ReturnChannel:
			s3Error = err
			if s3Error != nil {
				endLoop = true
			}
		case err := <-cacheRedirectionChannel.ReturnChannel:
			if err != nil {
				isFileCacheable = false
			}
		default:
			grpcMsg, err := uploadStream.Recv()
			if err == io.EOF {
				endLoop = true
				break
			}
			if err != nil {
				s3RedirectionChannel.ErrorChannel <- err
				if isFileCacheable {
					cacheRedirectionChannel.ErrorChannel <- err
				}
				endLoop = true
				grpcError = err
				break
			}

			s3Chunk := make([]byte, len(grpcMsg.Chunk))
			copy(s3Chunk, grpcMsg.Chunk)
			s3RedirectionChannel.ChunkChannel <- s3Chunk

			if isFileCacheable {
				cacheChunk := make([]byte, len(grpcMsg.Chunk))
				copy(cacheChunk, grpcMsg.Chunk)
				cacheRedirectionChannel.ChunkChannel <- cacheChunk
			}
		}
	}
	close(s3RedirectionChannel.ChunkChannel)
	close(s3RedirectionChannel.ErrorChannel)
	close(cacheRedirectionChannel.ChunkChannel)
	close(cacheRedirectionChannel.ErrorChannel)

	err := <-s3RedirectionChannel.ReturnChannel

	if err != nil {
		s3Error = err
	}
	if grpcError != nil {
		return fmt.Errorf("[*GRPC_ERROR*] - %s", grpcError.Error())
	} else if s3Error != nil {
		return fmt.Errorf("[*S3_ERROR*] - %s", s3Error.Error())
	} else {
		return nil
	}
}
