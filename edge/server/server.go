package server

import (
	"context"
	"edge/proto"
	"edge/utils"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
)

type Ticket struct {
	ServerEndpoint string
	Id             string
}

func publishTicket(channel *amqp.Channel, queueName string, ticket Ticket) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	encoded, err := json.Marshal(ticket)
	if err != nil {
		log.Println("[*ERROR*] - Error in marshaling Ticket for RabbitMQ")
		return err
	}
	err = channel.PublishWithContext(ctx, "", queueName, false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        encoded,
	})
	if err != nil {
		log.Println("[*ERROR*] - Error in publishing ticket on RabbitMQ")
		return err
	}
	return nil
}

func setupRabbitMQ(queueName string) (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial("amqp://guest:guest@rabbit_mq:5672/")
	if err != nil {
		log.Printf("[*ERROR*] - Impossibile contattare il server RabbitMQ\r\n")
		return nil, nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		log.Printf("[*ERROR*] - Impossibile aprire il canale verso la coda\r\n")
		return nil, nil, err
	}

	_, err = channel.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		log.Printf("[*ERROR*] - Impossibile dichiarare la coda\r\n")
		return nil, nil, err
	}

	return conn, channel, nil
}

var rabbitChannel *amqp.Channel

func ActAsServer() {
	serverEndpoint := setUpGRPC()
	queueName := utils.GetEnvironmentVariable("QUEUE_NAME")
	_, ch, err := setupRabbitMQ("storage_queue")
	rabbitChannel = ch
	utils.ExitOnError("[*ERROR*] - Errore sul setup della coda rabbit", err)
	//defer conn.Close()
	//defer ch.Close()
	for i := 0; i < utils.GetIntegerEnvironmentVariable("EDGE_TOKENS"); i++ {
		randomID, err := utils.GenerateRandomID()
		utils.ExitOnError("[*ERROR*] - Error generating random ID", err)
		err = publishTicket(rabbitChannel, queueName, Ticket{serverEndpoint, randomID})
	}
	utils.ExitOnError("[*ERROR*] - Impossibile pubblicare messaggio sulla coda\r\n", err)

	var forever chan struct{}

	log.Printf("[*] Waiting for messages. To exit press CTRL+C\r\n")
	<-forever
}

func setUpGRPC() string {
	ipAddr, err := utils.GetMyIPAddr()
	utils.ExitOnError(err.Error(), err)
	serverEndpoint := fmt.Sprintf("%s:%d", ipAddr, utils.GetRandomPort())
	lis, err := net.Listen("tcp", serverEndpoint)
	utils.ExitOnError("[*ERROR*] - failed to listen", err)
	grpcServer := grpc.NewServer()
	proto.RegisterFileServiceServer(grpcServer, &FileServiceServer{})
	log.Printf("[*] Waiting for requests on %s...", serverEndpoint)
	go grpcServer.Serve(lis)
	return serverEndpoint
}
