package server

import (
	"context"
	"edge/proto"
	"edge/utils"
	"encoding/json"
	"io"
	"log"
	"os"

	// Needed for handling X.509 certificates
	// Needed for reading files

	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type MethodType string

const (
	GET MethodType = "GET"
	PUT MethodType = "PUT"
	DEL MethodType = "DEL"
)

type Message struct {
	RequestId int32
	Method    MethodType
	FileName  string
	IpAddr    string
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

	go func() {
		for d := range msgs {
			var message Message
			json.Unmarshal(d.Body, &message)
			/*TLS CONFIGURATION ==============================================
			// Load CA certificate and key
			caCert, err := ioutil.ReadFile("/src/tls/ca-cert.pem")
			if err != nil {
				log.Fatalf("Failed to load CA certificate: %v", err)
			}
			// Create a certificate pool and add the CA certificate to it
			certPool := x509.NewCertPool()
			if !certPool.AppendCertsFromPEM(caCert) {
				log.Fatalf("failed to add server CA's certificate")
			}
			// Create a TLS configuration with the certificate pool
			tlsConfig := &tls.Config{
				RootCAs: certPool,
			}
			creds := credentials.NewTLS(tlsConfig)
			conn, err := grpc.Dial(message.IpAddr, grpc.WithTransportCredentials(creds))
			//================================================================*/
			conn, err := grpc.Dial(message.IpAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				log.Fatalf(err.Error())
			}
			log.Println(message)

			switch message.Method {

			case "GET":
				client := proto.NewFileServiceClient(conn)
				clientStream, err := client.Download(context.Background())
				if err != nil {
					log.Fatalf(err.Error())
				}
				// TODO Aggiungere ricerca all'interno della rete ed eventuale download del file da S3

				//SE CE L'HO IO:
				// Apri il file locale da cui verranno letti i chunks
				localFile, err := os.Open("/files/" + message.FileName)
				if err != nil {
					log.Fatalf(err.Error())
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
						log.Fatalf(err.Error())
					}
					clientStream.Send(&proto.FileChunk{RequestId: message.RequestId, FileName: message.FileName, Chunk: buffer[:n]})
				}

				response, err := clientStream.CloseAndRecv()
				if err != nil {
					log.Fatalf(err.Error())
				}
				if response.RequestId != message.RequestId {
					log.Printf("RequestID '%d' non riconosciuto! Expected --> '%d' ", response.RequestId, message.RequestId)
				} else if !response.Success {
					log.Printf("ERRORE nello scaricamento del File %s [REQ_ID: %d]", message.FileName, message.RequestId)
				} else {
					log.Printf("File %s caricato con successo [REQ_ID: %d]", message.FileName, message.RequestId)
				}
				break
			case "PUT":
				client := proto.NewFileServiceClient(conn)
				clientStream, err := client.Upload(context.Background(), &proto.FileUploadRequest{RequestId: message.RequestId, FileName: message.FileName})
				if err != nil {
					log.Fatalf(err.Error())
				}
				//il client invia il file --> l'Edge scarica i chunks (tramite questo stream)

				// Apri il file locale dove verranno scritti i chunks
				localFile, err := os.Create(message.FileName)
				if err != nil {
					log.Fatalf(err.Error())
				}
				defer localFile.Close()

				for {
					fileChunk, err := clientStream.Recv()
					if err == io.EOF {
						break
					}
					if err != nil {
						log.Fatalf(err.Error())
					}
					_, err = localFile.Write(fileChunk.Chunk)
					if err != nil {
						log.Fatalf(err.Error())
					}
				}

				log.Printf("File %s scaricato con successo", message.FileName)
				break
			}
			conn.Close()
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
