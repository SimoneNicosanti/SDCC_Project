package server

import (
	"edge/channels"
	"edge/proto/client"
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
				// Se c'è un errore sul canale di ridirezione principale (S3 oppure Client) il file non viene salvato in Cache
				// La copia creata viene quindi rimossa
				cacheRedirectionChannel.MessageChannel <- channels.Message{Body: []byte{}, Err: err}
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
				customErr := status.Error(codes.Code(client.ErrorCodes_CHUNK_ERROR), fmt.Sprintf("[*GRPC_ERROR*] - Failed while receiving chunks via gRPC.\r\nError: '%s'", err.Error()))
				mainRedirectionChannel.MessageChannel <- channels.Message{Body: []byte{}, Err: customErr}
				if isFileCacheable {
					cacheRedirectionChannel.MessageChannel <- channels.Message{Body: []byte{}, Err: customErr}
				}
				endLoop = true
				grpcError = err
				break
			}

			s3Chunk := make([]byte, len(grpcMsg.Chunk))
			copy(s3Chunk, grpcMsg.Chunk)
			mainRedirectionChannel.MessageChannel <- channels.Message{Body: s3Chunk, Err: nil}

			if isFileCacheable {
				cacheChunk := make([]byte, len(grpcMsg.Chunk))
				copy(cacheChunk, grpcMsg.Chunk)
				cacheRedirectionChannel.MessageChannel <- channels.Message{Body: cacheChunk, Err: nil}
			}
		}
	}
	close(mainRedirectionChannel.MessageChannel)
	close(cacheRedirectionChannel.MessageChannel)
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

func redirectStreamToClient(clientRedirectionChannel channels.RedirectionChannel, clientDownloadStream client.FileService_DownloadServer) error {
	defer close(clientRedirectionChannel.ReturnChannel)
	for message := range clientRedirectionChannel.MessageChannel {
		if message.Err != nil {
			return message.Err
		}
		err := clientDownloadStream.Send(&client.FileChunk{Chunk: message.Body})
		if err != nil {
			clientRedirectionChannel.ReturnChannel <- err
			return err
		}
	}
	return nil
}
