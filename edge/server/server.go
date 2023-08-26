package server

import (
	"context"
	"edge/proto"
	"edge/utils"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
)

type Ticket struct {
	ServerEndpoint string
	Id             string
}

type AuthorizedTicketIDs struct {
	mutex sync.RWMutex
	IDs   []string
}

var authorizedTicketIDs = AuthorizedTicketIDs{
	mutex: sync.RWMutex{},
	IDs:   make([]string, utils.GetIntegerEnvironmentVariable("EDGE_TICKETS_NUM")),
}

var serverEndpoint string

func attemptPublishTicket(channel *amqp.Channel, ticket Ticket) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	encoded, err := json.Marshal(ticket)
	if err != nil {
		log.Println("[*ERROR*] -> Error in marshaling Ticket for RabbitMQ")
		return err
	}
	err = channel.PublishWithContext(ctx, "", utils.GetEnvironmentVariable("QUEUE_NAME"), false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        encoded,
	})
	if err != nil {
		log.Println("[*ERROR*] -> Error in publishing ticket on RabbitMQ")
		return err
	}
	return nil
}

func publishTicket(channel *amqp.Channel, oldTicketIndex int) error {
	count := 0
	ticket := createTicket(oldTicketIndex)
	for count < 3 {
		err := attemptPublishTicket(channel, ticket)

		// La funzione ritorna al primo tentativo con successo
		if err == nil {
			return nil
		}
		count++
		log.Println(err.Error())
	}
	// Dopo tre tentativi falliti verrà generato un errore
	return fmt.Errorf("[*ERROR*] -> All the attempts to publish ticket '%s' failed", ticket.Id)
}

func createTicket(oldTicketIndex int) Ticket {
	randomID, err := utils.GenerateUniqueRandomID(authorizedTicketIDs.IDs)
	utils.ExitOnError("[*ERROR*] -> Error generating random ID for ticket", err)
	authorizedTicketIDs.IDs[oldTicketIndex] = randomID
	ticket := Ticket{serverEndpoint, randomID}
	log.Printf("[*TICKET GENERATED*] -> randomID : '%s'", randomID)
	return ticket
}

func setupRabbitMQ() *amqp.Channel {
	conn, err := amqp.Dial("amqp://guest:guest@rabbit_mq:5672/")
	utils.ExitOnError("[*ERROR*] -> Impossibile contattare il server RabbitMQ\r\n", err)

	channel, err := conn.Channel()
	utils.ExitOnError("[*ERROR*] -> Impossibile aprire il canale verso la coda\r\n", err)

	_, err = channel.QueueDeclare(utils.GetEnvironmentVariable("QUEUE_NAME"), true, false, false, false, nil)
	utils.ExitOnError("[*ERROR*] -> Impossibile dichiarare la coda\r\n", err)

	return channel
}

func publishAllTicketsOnQueue(rabbitChannel *amqp.Channel) {

	for i := 0; i < len(authorizedTicketIDs.IDs); i++ {
		err := publishTicket(rabbitChannel, i)
		utils.ExitOnError("[*ERROR*] -> Impossibile pubblicare ticket sulla coda\r\n", err)
	}
}

func ActAsServer() {
	setUpGRPC()

	rabbitChannel := setupRabbitMQ()

	publishAllTicketsOnQueue(rabbitChannel)

	var forever chan struct{}
	<-forever
}

func setUpGRPC() {
	ipAddr, err := utils.GetMyIPAddr()
	utils.ExitOnError("[*ERROR*] -> failed to retrieve server IP address", err)
	serverEndpoint = fmt.Sprintf("%s:%d", ipAddr, utils.GetRandomPort())
	lis, err := net.Listen("tcp", serverEndpoint)
	utils.ExitOnError("[*ERROR*] -> failed to listen", err)
	grpcServer := grpc.NewServer()
	proto.RegisterFileServiceServer(grpcServer, &FileServiceServer{})
	log.Printf("[*GRPC SERVER STARTED*] -> endpoint : '%s'", serverEndpoint)
	go grpcServer.Serve(lis)
}
