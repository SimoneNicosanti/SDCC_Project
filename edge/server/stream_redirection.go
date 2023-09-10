package server

import (
	"edge/channels"
	"edge/proto/client"
	"edge/utils"
	"fmt"
	"io"

	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

type GrpcReceiver interface {
	Recv() (*client.FileChunk, error)
}

func rcvAndRedirectChunks(mainRedirectionChannel channels.RedirectionChannel, cacheRedirectionChannel channels.RedirectionChannel, isFileCacheable bool, grpcReceiver GrpcReceiver) error {
	var endLoop bool = false
	var mainChannelError error = nil
	var grpcError error = nil

	for !endLoop {
		select {
		case err := <-mainRedirectionChannel.ReturnChannel:
			mainChannelError = err
			if mainChannelError != nil {
				endLoop = true
			}
		case err := <-cacheRedirectionChannel.ReturnChannel:
			if err != nil {
				isFileCacheable = false
			}
		default:
			grpcMsg, err := grpcReceiver.Recv()
			if err == io.EOF {
				endLoop = true
				break
			}
			if err != nil {
				mainRedirectionChannel.ErrorChannel <- status.Error(codes.Code(client.ErrorCodes_CHUNK_ERROR), fmt.Sprintf("[*GRPC_ERROR*] - Failed while receiving chunks via gRPC.\r\nError: '%s'", err.Error()))
				if isFileCacheable {
					cacheRedirectionChannel.ErrorChannel <- status.Error(codes.Code(client.ErrorCodes_CHUNK_ERROR), fmt.Sprintf("[*GRPC_ERROR*] - Failed while receiving chunks via gRPC.\r\nError: '%s'", err.Error()))
				}
				endLoop = true
				grpcError = err
				break
			}

			s3Chunk := make([]byte, len(grpcMsg.Chunk))
			copy(s3Chunk, grpcMsg.Chunk)
			mainRedirectionChannel.ChunkChannel <- s3Chunk

			if isFileCacheable {
				cacheChunk := make([]byte, len(grpcMsg.Chunk))
				copy(cacheChunk, grpcMsg.Chunk)
				cacheRedirectionChannel.ChunkChannel <- cacheChunk
			}
		}
	}
	close(mainRedirectionChannel.ChunkChannel)
	close(mainRedirectionChannel.ErrorChannel)
	close(cacheRedirectionChannel.ChunkChannel)
	close(cacheRedirectionChannel.ErrorChannel)

	err := <-mainRedirectionChannel.ReturnChannel

	if err != nil {
		mainChannelError = err
	}
	if grpcError != nil {
		return fmt.Errorf("[*GRPC_ERROR*] - %s", grpcError.Error())
	} else if mainChannelError != nil {
		return fmt.Errorf("[*MAIN_CHANNEL_ERROR*] - %s", mainChannelError.Error())
	} else {
		return nil
	}
}

func redirectStreamToClient(clientRedirectionChannel channels.RedirectionChannel, clientDownloadStream client.FileService_DownloadServer) {
	defer close(clientRedirectionChannel.ReturnChannel)
	var endLoop bool = false
	for !endLoop {
		select {
		case chunk, _ := <-clientRedirectionChannel.ChunkChannel:
			if len(chunk) == 0 {
				endLoop = true
				break
			}
			err := clientDownloadStream.Send(&client.FileChunk{Chunk: chunk})
			if err != nil {
				clientRedirectionChannel.ReturnChannel <- err
				endLoop = true
			}
		case err := <-clientRedirectionChannel.ErrorChannel:
			if err != nil {
				utils.PrintEvent("REDIRECTION_ERROR", err.Error())
				endLoop = true
				break
			}
		}
	}
	return
}
