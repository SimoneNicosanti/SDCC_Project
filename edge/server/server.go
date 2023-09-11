package server

import (
	"context"
	"edge/cache"
	"edge/proto/client"
	"edge/utils"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
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

var authorizedTicketIDs AuthorizedTicketIDs

var serverEndpoint string
var rabbitChannel *amqp.Channel

func attemptPublishTicket(channel *amqp.Channel, ticket Ticket) error {
	encoded, err := json.Marshal(ticket)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = channel.PublishWithContext(ctx, "", utils.GetEnvironmentVariable("QUEUE_NAME"), false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        encoded,
	})

	return err
}

func publishNewTicket(oldTicketIndex int) {
	count := 0
	ticket := createTicket(oldTicketIndex)
	for count < 3 {
		err := attemptPublishTicket(rabbitChannel, ticket)
		// La funzione ritorna al primo tentativo con successo
		if err == nil {
			utils.PrintEvent("TICKET_PUBLISHED", fmt.Sprintf("il ticket '%s' è stato pubblicato", ticket.Id))
			return
		} else {
			if strings.Contains(err.Error(), "Exception (504) Reason: \"channel/connection is not open\"") { //TODO da controllare
				setupRabbitMQ()
			}
		}
		count++
		utils.PrintEvent("RABBITMQ_ERROR", fmt.Sprintf("Impossibile pubblicare ticket '%s' su rabbitMQ per la %d volta.\r\nL'errore restituito è: '%s'", ticket.Id, count, err.Error()))
	}
	// Dopo tre tentativi falliti verrà generato un errore
	utils.PrintEvent("RABBITMQ_ERROR", fmt.Sprintf("Tutti i tentativi di pubblicare il ticket '%s' non hanno avuto successo", ticket.Id))
	//TODO errore fatale!!
	log.Panic("FATAL_ERR -> rabbitmq fatal err")

}

func createTicket(oldTicketIndex int) Ticket {
	randomID, err := utils.GenerateUniqueRandomID(authorizedTicketIDs.IDs)
	utils.ExitOnError("[*ERROR*] -> Impossibile generare un ID random", err)
	authorizedTicketIDs.IDs[oldTicketIndex] = randomID
	ticket := Ticket{serverEndpoint, randomID}
	utils.PrintEvent("TICKET_GENERATED", "il ticket '"+randomID+"' è stato generato")
	return ticket
}

func setupRabbitMQ() {
	conn, err := amqp.Dial("amqp://guest:guest@rabbit_mq:5672/")
	utils.ExitOnError("[*ERROR*] -> Impossibile contattare il server RabbitMQ\r\n", err)

	rabbitChannel, err = conn.Channel()
	utils.ExitOnError("[*ERROR*] -> Impossibile aprire il canale verso la coda\r\n", err)

	_, err = rabbitChannel.QueueDeclare(utils.GetEnvironmentVariable("QUEUE_NAME"), true, false, false, false, nil)
	utils.ExitOnError("[*ERROR*] -> Impossibile dichiarare la coda\r\n", err)
}

func publishAllTicketsOnQueue(rabbitChannel *amqp.Channel) {
	authorizedTicketIDs = AuthorizedTicketIDs{
		mutex: sync.RWMutex{},
		IDs:   make([]string, utils.GetIntEnvironmentVariable("EDGE_TICKETS_NUM")),
	}
	for i := 0; i < len(authorizedTicketIDs.IDs); i++ {
		publishNewTicket(i)
	}
}

func ActAsServer() {
	cache.GetCache().ActivateCacheRecovery()
	setUpGRPC()
	setupRabbitMQ()
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
	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(utils.GetIntEnvironmentVariable("MAX_GRPC_MESSAGE_SIZE")), // Imposta la nuova dimensione massima
	}
	grpcServer := grpc.NewServer(opts...)
	client.RegisterFileServiceServer(grpcServer, &FileServiceServer{})
	utils.PrintEvent("GRPC_SERVER_STARTED", "Server Endpoint : "+serverEndpoint)
	go grpcServer.Serve(lis)
}
