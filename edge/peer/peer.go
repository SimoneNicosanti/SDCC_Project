package peer

import (
	"bytes"
	"edge/utils"
	"errors"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"strings"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	// boom "github.com/tylertreat/BoomFilters"
)

/*
	TODO : Azione nell'heartbeat (Peer -> Registry)
	TODO : Ricezione dei ping -> se ricevo un ping/richiesta di qualcuno che non conosco, allora lo inserisco come vicino
	Logica di rilevamento errori: heartbeat verso il registry e ping verso i vicini.
	In caso di caduta del registry, questo può richiedere ai peer di comunicargli i loro vicini in modo tale che il registry
	può ricreare la rete.
	TODO : Thread che risponde alle richieste nella coda Rabbit_mq
	TODO : Meccanismo di caching
*/

type EdgePeer struct {
	PeerAddr string
}

var selfPeer EdgePeer

func ActAsPeer() {
	// bloomFilter := boom.NewDefaultStableBloomFilter(10000, 0.01)
	// fmt.Println(bloomFilter)

	addresses, _ := net.InterfaceAddrs()
	ipAddr := strings.Split(addresses[1].String(), "/")[0]

	//Registrazione del servizio e lancio di un thread in ascolto
	edgePeerPtr := new(EdgePeer)
	errorMessage, err := registerServiceForEdge(ipAddr, edgePeerPtr)
	utils.ExitOnError(errorMessage, err)
	log.Println("Servizio registrato")

	//Connessione al server Registry per l'inserimento nella rete
	adj := new([]EdgePeer)
	registryClientPtr, errorMessage, err := registerToRegistry(edgePeerPtr, adj)
	utils.ExitOnError("Impossibile registrare il servizio sul registry server: "+errorMessage, err)
	log.Println("Servizio registrato su server Registry")

	log.Println(*adj) //Vicini restituiti dal server Registry

	startHeartBeatThread() //Inizio meccanismo di heartbeat verso il server Registry
	log.Println("Heartbeat iniziato")

	//Connessione a tutti i vicini
	connectAndNotifyYourAdjacent(*adj)
	log.Println("Connessione con tutti i vicini completata")

	defer registryClientPtr.Close()

	var forever chan struct{}
	<-forever
}

//host='rabbit_mq', port = '5672'
func consumeMessages() {
	conn, err := amqp.Dial("amqp://guest:guest@rabbit_mq:5672/")
	utils.ExitOnError("Failed to connect to RabbitMQ", err)
	defer conn.Close()

	ch, err := conn.Channel()
	utils.ExitOnError("Failed to open a channel", err)
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"storage_queue", // name
		true,            // durable
		false,           // delete when unused
		false,           // exclusive
		false,           // no-wait
		nil,             // arguments
	)
	utils.ExitOnError("Failed to declare a queue", err)

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	utils.ExitOnError("Failed to set QoS", err)

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	utils.ExitOnError("Failed to register as a consumer", err)

	var forever chan struct{}

	go func() {
		for d := range msgs {
			log.Printf("[*] Received a message: %s", d.Body)
			dotCount := bytes.Count(d.Body, []byte("."))
			t := time.Duration(dotCount)
			time.Sleep(t * time.Second)
			log.Printf("[*] Done")
			d.Ack(false)
		}
	}()

	log.Printf("[*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func connectAndNotifyYourAdjacent(adj []EdgePeer) {
	for i := 0; i < len(adj); i++ {
		client, err := connectAndAddNeighbour(adj[i])
		//Nel caso in cui uno dei vicini non rispondesse alla nostra richiesta di connessione,
		// il peer corrente lo ignorerà.
		if err != nil {
			continue
		}

		log.Println("Connessione con " + adj[i].PeerAddr + " effettuata")

		err = CallAdjAddNeighbour(client, selfPeer)
		//Se il vicino a cui ci si è connessi non ricambia la connessione, chiudo la connessione stabilita precedentemente.
		if err != nil {
			client.Close()
			log.Println(err.Error())
			continue
		}
	}
}

func connectToPeer(addr string) (*rpc.Client, string, error) {
	client, err := rpc.DialHTTP("tcp", addr)
	if err != nil {
		return nil, "Errore Dial HTTP", err
	}
	return client, "", err
}

func registerToRegistry(edgePeerPtr *EdgePeer, adj *[]EdgePeer) (*rpc.Client, string, error) {
	registryAddr := "registry:1234"
	client, err := rpc.DialHTTP("tcp", registryAddr)
	if err != nil {
		return nil, "Errore Dial HTTP", err
	}

	RegistryConn = PeerConnection{EdgePeer{registryAddr}, client}

	err = client.Call("RegistryService.PeerEnter", *edgePeerPtr, adj)
	if err != nil {
		return nil, "Errore durante la registrazione al Registry Server", err
	}

	return client, "", nil
}

func registerServiceForEdge(ipAddrStr string, edgePeerPtr *EdgePeer) (string, error) {
	err := rpc.Register(edgePeerPtr)
	if err != nil {
		return "Errore registrazione del servizio", err
	}

	rpc.HandleHTTP()
	bindIpAddr := ipAddrStr + ":0"

	peerListener, err := net.Listen("tcp", bindIpAddr)
	if err != nil {
		return "Errore listen", err
	}
	edgePeerPtr.PeerAddr = peerListener.Addr().String()

	//Thread che ascolta eventuali richieste arrivate
	go listenLoop(peerListener)

	selfPeer = EdgePeer{edgePeerPtr.PeerAddr}

	return "", err
}

func listenLoop(listener net.Listener) {
	for {
		http.Serve(listener, nil)
	}
}

func startHeartBeatThread() {
	hb, err := utils.GetConfigFieldFromFile("conf.properties", "HEARTBEAT_FREQUENCY")
	if err != nil {
		log.Println("Error retreving HEARTBEAT_FREQUENCY from properties", err)
		return
	}
	heartbeatDuration, err := time.ParseDuration(hb + "s")
	if err != nil {
		log.Println("Error parsing heartbeat duration:", err)
		return
	}
	// Start the heartbeat thread
	go heartbeatToRegistry(heartbeatDuration)
}

func heartbeatToRegistry(interval time.Duration) {
	for {
		// TODO Perform heartbeat action here
		log.Println("Heartbeat action executed.")
		// Wait for the specified interval before the next heartbeat
		time.Sleep(interval)
	}
}

func connectAndAddNeighbour(peer EdgePeer) (*rpc.Client, error) {
	client, errorMessage, err := connectToPeer(peer.PeerAddr)
	//Nel caso in cui uno dei vicini non rispondesse alla nostra richiesta di connessione,
	// il peer corrente lo ignorerà.
	if err != nil {
		log.Println(errorMessage + "Impossibile stabilire la connessione con " + peer.PeerAddr)
		return nil, errors.New(errorMessage + "Impossibile stabilire la connessione con " + peer.PeerAddr)
	}

	//Connessione con il vicino creata correttamente, quindi la aggiungiamo al nostro insieme di connessioni
	peerConn := PeerConnection{peer, client}

	AddConnection(peerConn)

	return client, nil
}
