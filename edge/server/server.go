package server

import (
	"edge/utils"
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
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
		false,           // durable
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
			// fmt.Println(d.Body)
			log.Println(message.Method, message.FileName)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
