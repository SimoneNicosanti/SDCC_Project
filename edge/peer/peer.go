package peer

import (
	"edge/cache"
	"edge/proto/client"
	"edge/utils"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/rpc"
	"time"

	"google.golang.org/grpc"
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

var selfPeer EdgePeer
var peerFileServer PeerFileServer
var registryClient *rpc.Client

func ActAsPeer() {
	// bloomFilter := boom.NewDefaultStableBloomFilter(10000, 0.01)
	// fmt.Println(bloomFilter)
	ipAddr, err := utils.GetMyIPAddr()
	utils.ExitOnError("", err)

	//setupBloomFilterStruct()

	//Registrazione del servizio e lancio di un thread in ascolto
	edgePeerPtr := new(EdgePeer)
	errorMessage, err := registerServiceForEdge(ipAddr, edgePeerPtr)
	utils.ExitOnError(errorMessage, err)
	log.Println("[*OK*] -> Servizio registrato")

	registerGRPC()

	//Connessione al server Registry per l'inserimento nella rete
	adj := new(map[EdgePeer]byte)
	registryClientPtr, errorMessage, err := registerToRegistry(edgePeerPtr, adj)
	registryClient = registryClientPtr

	utils.ExitOnError("[*ERROR*] -> Impossibile registrare il servizio sul registry server: "+errorMessage, err)
	log.Println("[*OK*] -> Servizio registrato su server Registry")

	log.Println(*adj) //Vicini restituiti dal server Registry

	go heartbeatToRegistry() //Inizio meccanismo di heartbeat verso il server Registry
	log.Println("[*HEARTBEAT*] -> iniziato")

	//TODO Far partire il meccanismo di ping e di notifica dei filtr
	go temporizedNotifyBloomFilters()

	//Connessione a tutti i vicini
	connectAndNotifyYourAdjacent(*adj)
	log.Println("[*NETWORK*] -> Connessione con tutti i vicini completata")

	//defer registryClientPtr.Close()

}

// Registra il server gRPC per ricevere le richieste di trasferimento file dagli edge
func registerGRPC() {
	ipAddr, err := utils.GetMyIPAddr()
	utils.ExitOnError("[*ERROR*] -> failed to retrieve server IP address", err)

	serverEndpoint := fmt.Sprintf("%s:%d", ipAddr, utils.GetRandomPort())
	lis, err := net.Listen("tcp", serverEndpoint)
	utils.ExitOnError("[*ERROR*] -> failed to listen", err)

	peerFileServer.IpAddr = serverEndpoint

	grpcServer := grpc.NewServer()
	client.RegisterEdgeFileServiceServer(grpcServer, &peerFileServer)

	log.Printf("[*GRPC SERVER STARTED*] -> endpoint : '%s'", selfPeer.PeerAddr)
	go grpcServer.Serve(lis)
}

func connectAndNotifyYourAdjacent(adjs map[EdgePeer]byte) {
	for adjPeer := range adjs {
		client, err := connectAndAddNeighbour(adjPeer)
		//Nel caso in cui uno dei vicini non rispondesse alla nostra richiesta di connessione,
		// il peer corrente lo ignorerà.
		if err != nil {
			continue
		}

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

	client, err := ConnectToNode(registryAddr)
	utils.ExitOnError("", err)

	err = client.Call("RegistryService.PeerEnter", *edgePeerPtr, adj)
	if err != nil {
		return nil, "[*ERROR*] -> Errore durante la registrazione al Registry Server", err
	}

	return client, "", nil
}

func registerServiceForEdge(ipAddrStr string, edgePeerPtr *EdgePeer) (string, error) {
	err := rpc.Register(edgePeerPtr)
	if err != nil {
		return "[*ERROR*] -> Errore registrazione del servizio", err
	}

	rpc.HandleHTTP()
	bindIpAddr := ipAddrStr + ":0"

	peerListener, err := net.Listen("tcp", bindIpAddr)
	if err != nil {
		return "[*ERROR*] -> Errore listen", err
	}
	edgePeerPtr.PeerAddr = peerListener.Addr().String()

	// Thread che ascolta eventuali richieste arrivate
	go http.Serve(peerListener, nil)

	selfPeer = EdgePeer{PeerAddr: edgePeerPtr.PeerAddr}

	return "", err
}

func temporizedNotifyBloomFilters() {
	FILTER_NOTIFY_TIME := utils.GetIntegerEnvironmentVariable("FILTER_NOTIFY_TIME")
	for {
		time.Sleep(time.Duration(FILTER_NOTIFY_TIME) * time.Second)
		notifyBloomFilters()
	}
}

func notifyBloomFilters() error {
	adjacentsMap.connsMutex.RLock()
	adjacentsMap.filtersMutex.RLock()

	defer adjacentsMap.filtersMutex.RUnlock()
	defer adjacentsMap.connsMutex.RUnlock()

	bloomFilter := cache.GetCache().ComputeBloomFilter()
	filterEncode, err := bloomFilter.GobEncode()
	if err != nil {
		return err
	}
	for edgePeer, adjConn := range adjacentsMap.peerConns {
		filterMessage := BloomFilterMessage{EdgePeer: selfPeer, BloomFilter: filterEncode}
		err := adjConn.Call("EdgePeer.NotifyBloomFilter", filterMessage, new(int))
		if err != nil {
			log.Println("[*ERROR*] -> Impossiile notificare il Filtro di Bloom a " + edgePeer.PeerAddr)
			return err
		}
	}

	return nil

}

func NeighboursFileLookup(fileRequestMessage FileRequestMessage) (EdgePeer, error) {
	adjacentsMap.connsMutex.RLock()
	adjacentsMap.filtersMutex.RLock()

	defer adjacentsMap.filtersMutex.RUnlock()
	defer adjacentsMap.connsMutex.RUnlock()

	doneChannel := make(chan *rpc.Call)
	maxContactable := utils.GetIntegerEnvironmentVariable("MAX_CONTACTABLE_ADJ")
	contactedNum := 0
	for adj := range adjacentsMap.peerConns {
		adjFilter, isInMap := adjacentsMap.filterMap[adj]
		if isInMap {
			if adjFilter.Test([]byte(fileRequestMessage.FileName)) {
				contactNeighbourForFile(fileRequestMessage, adj, doneChannel)
				contactedNum++
				if contactedNum == maxContactable {
					break
				}
			}
		}
	}

	if contactedNum < maxContactable {
		falseFiltersNeighbours := findFalseAdjacentsFilter(fileRequestMessage.FileName)
		for i := 0; i < len(falseFiltersNeighbours); i++ {
			randomInt := rand.Intn(len(falseFiltersNeighbours))
			randomNeigh := falseFiltersNeighbours[randomInt]
			contactNeighbourForFile(fileRequestMessage, randomNeigh, doneChannel)
			contactedNum++
			if contactedNum == maxContactable {
				break
			}
			falseFiltersNeighbours = append(falseFiltersNeighbours[0:randomInt], falseFiltersNeighbours[randomInt+1:]...)
		}
	}

	for i := 0; i < contactedNum; i++ {
		neighbourCall := <-doneChannel
		err := neighbourCall.Error
		if err == nil {
			return neighbourCall.Reply.(EdgePeer), nil
		}
	}

	return EdgePeer{}, fmt.Errorf("[*ERROR*] -> No Neighbour Answered Successfully to Lookup for file %s", fileRequestMessage.FileName)

}

func findFalseAdjacentsFilter(fileName string) []EdgePeer {
	falseAdjacentsList := make([]EdgePeer, 0)
	for adj, adjFilter := range adjacentsMap.filterMap {
		if !adjFilter.Test([]byte(fileName)) {
			falseAdjacentsList = append(falseAdjacentsList, adj)
		}
	}
	return falseAdjacentsList
}

func contactNeighbourForFile(fileRequestMessage FileRequestMessage, adj EdgePeer, doneChannel chan *rpc.Call) {
	ownerPeerPtr := new(EdgePeer)
	adjConn := adjacentsMap.peerConns[adj]
	adjConn.Go("EdgePeer.FileLookup", fileRequestMessage, ownerPeerPtr, doneChannel)
}
