package peer

import (
	//"context"
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

// TODO : Ricezione dei ping -> se ricevo un ping/richiesta di qualcuno che non conosco, allora lo inserisco come vicino
// Logica di rilevamento errori: heartbeat verso il registry e ping verso i vicini.
// In caso di caduta del registry, questo può richiedere ai peer di comunicargli i loro vicini in modo tale che il registry
// può ricreare la rete.

var SelfPeer EdgePeer
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
	utils.PrintEvent("SERVER_SERVICE_OK", "Servizio registrato correttamente")

	registerGRPC()

	//Connessione al server Registry per l'inserimento nella rete
	adj := new(map[EdgePeer]byte)
	registryClientPtr, errorMessage, err := registerToRegistry(edgePeerPtr, adj)
	registryClient = registryClientPtr

	utils.ExitOnError("[*ERROR*] -> Impossibile registrare il servizio sul registry server: "+errorMessage, err)
	utils.PrintEvent("EDGE_SERVICE_OK", "Servizio registrato su server Registry")

	utils.PrintEvent("NEIGHBOURS_RECEIVED", "Vicini restituiti dal server Registry:\r\n"+fmt.Sprintln(*adj)) //Vicini restituiti dal server Registry

	go heartbeatToRegistry() //Inizio meccanismo di heartbeat verso il server Registry
	utils.PrintEvent("HEARTBEAT", "Inizio meccanismo di heartbeat verso il server Registry")

	//TODO Far partire il meccanismo di ping
	go temporizedNotifyBloomFilters()

	//Connessione a tutti i vicini
	connectAndNotifyYourAdjacent(*adj)
	utils.PrintEvent("NETWORK_COMPLETED", "Connessione con tutti i vicini completata")

	//defer registryClientPtr.Close()

}

// Registra il server gRPC per ricevere le richieste di trasferimento file dagli edge
func registerGRPC() {
	ipAddr, err := utils.GetMyIPAddr()
	utils.ExitOnError("[*ERROR*] -> failed to retrieve server IP address", err)

	serverEndpoint := fmt.Sprintf("%s:%d", ipAddr, utils.GetRandomPort())
	lis, err := net.Listen("tcp", serverEndpoint)
	utils.ExitOnError("[*ERROR*] -> failed to listen", err)

	peerFileServer = PeerFileServer{IpAddr: serverEndpoint}
	//peerFileServer.IpAddr = serverEndpoint

	grpcServer := grpc.NewServer()
	client.RegisterEdgeFileServiceServer(grpcServer, &peerFileServer)

	utils.PrintEvent("GRPC_SERVER_STARTED", "Server Endpoint : "+SelfPeer.PeerAddr)
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

		err = CallAdjAddNeighbour(client, SelfPeer)
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

	SelfPeer = EdgePeer{PeerAddr: edgePeerPtr.PeerAddr}

	return "", err
}

func temporizedNotifyBloomFilters() {
	FILTER_NOTIFY_TIME := utils.GetIntEnvironmentVariable("FILTER_NOTIFY_TIME")
	for {
		time.Sleep(time.Duration(FILTER_NOTIFY_TIME) * time.Second)
		notifyBloomFiltersToAdjacents()
	}
}

func notifyBloomFiltersToAdjacents() error {
	adjacentsMap.connsMutex.RLock()
	adjacentsMap.filtersMutex.RLock()

	defer adjacentsMap.filtersMutex.RUnlock()
	defer adjacentsMap.connsMutex.RUnlock()

	bloomFilter := cache.GetCache().ComputeBloomFilter()
	filterEncode, err := bloomFilter.GobEncode()
	if err != nil {
		return err
	}

	// ctx, cancel := context.WithTimeout(context.Background(), time.Duration(utils.GetIntegerEnvironmentVariable("MAX_WAITING_TIME_FOR_CALL")))
	for edgePeer, adjConn := range adjacentsMap.peerConns {
		filterMessage := BloomFilterMessage{EdgePeer: SelfPeer, BloomFilter: filterEncode}
		err := adjConn.Call("EdgePeer.NotifyBloomFilter", filterMessage, new(int))
		//context.WithTimeout()
		if err != nil {
			utils.PrintEvent("BLOOM_ERROR", "impossibile inviare filtro a "+edgePeer.PeerAddr+".\r\nL'errore restituito è: '"+err.Error()+"'.")
			return err
		}
	}

	return nil

}

func NeighboursFileLookup(fileRequestMessage FileRequestMessage) (FileLookupResponse, error) {
	adjacentsMap.connsMutex.RLock()
	adjacentsMap.filtersMutex.RLock()

	defer adjacentsMap.filtersMutex.RUnlock()
	defer adjacentsMap.connsMutex.RUnlock()

	fileRequestMessage.ForwarderPeer = SelfPeer

	maxContactable := utils.GetIntEnvironmentVariable("MAX_CONTACTABLE_ADJ")
	doneChannel := make(chan *rpc.Call, maxContactable)
	contactedNum := 0

	// Contattiamo solo i vicini positivi ai filtri (tranne il mittente originario)
	for adj := range adjacentsMap.peerConns {
		adjFilter, isInMap := adjacentsMap.filterMap[adj]
		if isInMap {
			if adjFilter.Test([]byte(fileRequestMessage.FileName)) {
				contacted := contactNeighbourForFile(fileRequestMessage, adj, doneChannel)
				if contacted {
					utils.PrintEvent("LOOKUP", "Richiesta inviata a "+adj.PeerAddr+" per filtro di Bloom")
					contactedNum++
					if contactedNum == maxContactable {
						break
					}
				}
			}
		}
	}

	// Contattiamo (come rimanenti) alcuni vicini a caso con i filtri negativi (tranne il mittente originario)
	if contactedNum < maxContactable {
		falseFiltersNeighbours := findFalseAdjacentsFilter(fileRequestMessage.FileName)
		for i := 0; i < (maxContactable - contactedNum); i++ {
			if len(falseFiltersNeighbours) == 0 {
				break
			}
			randomInt := rand.Intn(len(falseFiltersNeighbours))
			randomNeigh := falseFiltersNeighbours[randomInt]
			contacted := contactNeighbourForFile(fileRequestMessage, randomNeigh, doneChannel)
			if contacted {
				utils.PrintEvent("LOOKUP", "Richiesta inviata a "+randomNeigh.PeerAddr+" per complemento")
				contactedNum++
			}
			falseFiltersNeighbours = append(falseFiltersNeighbours[0:randomInt], falseFiltersNeighbours[randomInt+1:]...)
		}
	}

	// Ritorniamo il canale anziché il primo che risponde: quello che risponde potrebbe aver tolto il file dalla cache nel mentre
	for i := 0; i < contactedNum; i++ {
		neighbourCall := <-doneChannel
		err := neighbourCall.Error
		fileLookupResponsePtr := neighbourCall.Reply.(*FileLookupResponse)
		if err == nil {
			utils.PrintEvent("LOOKUP_RESPONSE_OK", "Riscontro positivo da "+fileLookupResponsePtr.OwnerEdge.IpAddr)
			return *fileLookupResponsePtr, nil
		}
	}
	return FileLookupResponse{}, fmt.Errorf("[*LOOKUP_ERROR*] -> No Neighbour Answered Successfully to Lookup for file %s", fileRequestMessage.FileName)

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

func contactNeighbourForFile(fileRequestMessage FileRequestMessage, adj EdgePeer, doneChannel chan *rpc.Call) bool {
	if adj != fileRequestMessage.SenderPeer {
		fileLookupResponsePtr := new(FileLookupResponse)
		adjConn := adjacentsMap.peerConns[adj]
		adjConn.Go("EdgePeer.FileLookup", fileRequestMessage, fileLookupResponsePtr, doneChannel)
		return true
	}
	return false
}
