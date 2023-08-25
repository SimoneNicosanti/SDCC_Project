package peer

import (
	"edge/utils"
	"log"
	"net"
	"net/http"
	"net/rpc"
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
var registryClient *rpc.Client

func ActAsPeer() {
	// bloomFilter := boom.NewDefaultStableBloomFilter(10000, 0.01)
	// fmt.Println(bloomFilter)
	ipAddr, err := utils.GetMyIPAddr()
	utils.ExitOnError(err.Error(), err)

	//Registrazione del servizio e lancio di un thread in ascolto
	edgePeerPtr := new(EdgePeer)
	errorMessage, err := registerServiceForEdge(ipAddr, edgePeerPtr)
	utils.ExitOnError(errorMessage, err)
	log.Println("Servizio registrato")

	//Connessione al server Registry per l'inserimento nella rete
	adj := new(map[EdgePeer]byte)
	registryClientPtr, errorMessage, err := registerToRegistry(edgePeerPtr, adj)
	registryClient = registryClientPtr

	utils.ExitOnError("Impossibile registrare il servizio sul registry server: "+errorMessage, err)
	log.Println("Servizio registrato su server Registry")

	log.Println(*adj) //Vicini restituiti dal server Registry

	go heartbeatToRegistry() //Inizio meccanismo di heartbeat verso il server Registry
	log.Println("Heartbeat iniziato")

	//Connessione a tutti i vicini
	connectAndNotifyYourAdjacent(*adj)
	log.Println("Connessione con tutti i vicini completata")

	//defer registryClientPtr.Close()

}

func connectAndNotifyYourAdjacent(adjs map[EdgePeer]byte) {
	for adjPeer := range adjs {
		client, err := connectAndAddNeighbour(adjPeer)
		//Nel caso in cui uno dei vicini non rispondesse alla nostra richiesta di connessione,
		// il peer corrente lo ignorerà.
		if err != nil {
			continue
		}

		log.Println("Connessione con " + adjPeer.PeerAddr + " effettuata")

		err = CallAdjAddNeighbour(client, selfPeer)
		//Se il vicino a cui ci si è connessi non ricambia la connessione, chiudo la connessione stabilita precedentemente.
		if err != nil {
			client.Close()
			log.Println(err.Error())
			continue
		}
	}
}

func registerToRegistry(edgePeerPtr *EdgePeer, adj *map[EdgePeer]byte) (*rpc.Client, string, error) {
	registryAddr := "registry:1234"
	client, err := rpc.DialHTTP("tcp", registryAddr)
	if err != nil {
		return nil, "Errore Dial HTTP", err
	}

	client, errorMessage, err := connectToNode("registry:1234")
	utils.ExitOnError(errorMessage, err)

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
