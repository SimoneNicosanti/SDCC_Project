package server

import (
	"context"
	"edge/proto"
	"edge/utils"
	"encoding/json"
	"io"
	"log"
	"os"

	"server/proto"

	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
)

type MethodType string

const (
	GET MethodType = "GET"
	PUT MethodType = "PUT"
	DEL MethodType = "DEL"
)

type Message struct {
	Method   MethodType
	FileName string
	IpAddr   string
}

func ActAsServer() {
	conn, err := amqp.Dial("amqp://guest:guest@rabbit_mq:5672/")
	utils.ExitOnError("Impossibile contattare il server RabbitMQ", err)
	defer conn.Close()

	ch, err := conn.Channel()
	utils.ExitOnError("Impossibile aprire il canale verso la coda", err)
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"storage_queue", // name
		true,            // durable
		false,           // delete when unused
		false,           // exclusive
		false,           // no-wait
		nil,             // arguments
	)
	utils.ExitOnError("Impossibile dichiarare la coda", err)

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		utils.ExitOnError("Impossibile registrare un consumer sulla coda", err)
	}

	var forever chan struct{}

	go func() (string, error) {
		for d := range msgs {
			var message Message
			json.Unmarshal(d.Body, &message)
			conn, err := grpc.Dial(message.IpAddr)
			if err != nil {
				return "Errore Dial grpc", err
			}

			switch message.Method {

			case "GET":
				client := proto.NewFileServiceClient(conn)
				clientStream, err := client.Download(context.Background())
				if err != nil {
					return "Errore durante tentativo di elaborazione GET request", err
				}
				// TODO Aggiungere ricerca all'interno della rete ed eventuale download del file da S3

				//SE CE L'HO IO:
				// Apri il file locale da cui verranno letti i chunks
				localFile, err := os.Open(message.FileName)
				if err != nil {
					return "Errore durante l'apertura del file locale", err
				}
				defer localFile.Close()
				chunkSize := 1024 // dimensione del chunk
				buffer := make([]byte, chunkSize)
				for {
					n, err := localFile.Read(buffer)
					if err == io.EOF {
						break
					}
					if err != nil {
						return "Errore durante la lettura del chunk dal file locale", err
					}
					clientStream.Send(&proto.FileChunk{Chunk: buffer[:n]})
				}

				_, err = clientStream.CloseAndRecv()
				if err != nil {
					return "Errore durante la conferma di upload del file", err
				}
				log.Printf("File %s caricato con successo", message.FileName)

			case "PUT":
				client := proto.NewFileServiceClient(conn)
				clientStream, err := client.Upload(context.Background(), &proto.FileUploadRequest{FileName: message.FileName})
				if err != nil {
					return "Errore durante tentativo di elaborazione PUT request", err
				}
				//il client invia il file --> l'Edge scarica i chunks (tramite questo stream)

				// Apri il file locale dove verranno scritti i chunks
				localFile, err := os.Create(message.FileName)
				if err != nil {
					return "Errore durante l'apertura del file locale", err
				}
				defer localFile.Close()

				for {
					fileChunk, err := clientStream.Recv()
					if err == io.EOF {
						break
					}
					if err != nil {
						return "Errore durante la ricezione del chunk", err
					}
					_, err = localFile.Write(fileChunk.Chunk)
					if err != nil {
						return "Errore durante la scrittura del chunk nel file locale", err
					}
				}

				log.Printf("File %s scaricato con successo", message.FileName)
			}
			conn.Close()
			// fmt.Println(d.Body)
			log.Println(message.Method, message.FileName)
		}

		return "", nil
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
