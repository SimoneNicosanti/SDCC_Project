package peer

import (
	"edge/cache"
	"edge/proto/file_transfer"
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

var SelfPeer EdgePeer
var peerFileServer PeerFileServer
var registryClient *rpc.Client

func ActAsPeer() {
	ipAddr, err := utils.GetMyIPAddr()
	utils.ExitOnError("", err)

	//Registrazione del servizio e lancio di un thread in ascolto
	edgePeerPtr := new(EdgePeer)
	errorMessage, err := registerServiceForEdge(ipAddr, edgePeerPtr)
	utils.ExitOnError(errorMessage, err)
	utils.PrintEvent("SERVER_SERVICE_OK", "Servizio registrato correttamente")

	registerGRPC()

	//Connessione al server Registry per l'inserimento nella rete
	adj := new(map[EdgePeer]byte)
	registryClientPtr, err := registerToRegistry(edgePeerPtr, adj)
	registryClient = registryClientPtr

	utils.ExitOnError("[*ERROR*] -> Impossibile registrare il servizio sul registry server", err)
	utils.PrintEvent("EDGE_SERVICE_OK", "Servizio registrato su server Registry")
	stringAdjMap := make(map[string]byte)
	for peer := range *adj {
		stringAdjMap[peer.PeerAddr] = 0
	}
	utils.PrintCustomMap(stringAdjMap, "Nessun vicino a cui connettersi...", "Vicini restituiti dal server Registry", "NEIGHBOURS_RECEIVED")

	go heartbeatToRegistry() //Inizio meccanismo di heartbeat verso il server Registry
	go pingsToAdjacents()
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

	serverEndpoint := ipAddr + ":0"
	lis, err := net.Listen("tcp", serverEndpoint)
	//Otteniamo l'indirizzo usato
	serverEndpoint = lis.Addr().String()
	utils.ExitOnError("[*ERROR*] -> failed to listen", err)

	opts := []grpc.ServerOption{
		grpc.MaxSendMsgSize(utils.GetIntEnvironmentVariable("MAX_GRPC_MESSAGE_SIZE")), // Imposta la nuova dimensione massima
	}
	peerFileServer = PeerFileServer{IpAddr: serverEndpoint}
	grpcServer := grpc.NewServer(opts...)
	file_transfer.RegisterEdgeFileServiceServer(grpcServer, &peerFileServer)

	utils.PrintEvent("GRPC_SERVER_STARTED", "Grpc server started in peer with endpoint : "+SelfPeer.PeerAddr)
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

func registerToRegistry(edgePeerPtr *EdgePeer, adj *map[EdgePeer]byte) (*rpc.Client, error) {
	registryAddr := "registry:1234"

	client, err := utils.ConnectToNode(registryAddr)
	utils.ExitOnError("", err)

	call := client.Go("RegistryService.PeerEnter", *edgePeerPtr, adj, nil)
	select {
	case <-call.Done:
		if call.Error != nil {
			return nil, fmt.Errorf("[*ERROR*] -> Errore durante la registrazione al Registry Server. L'errore restituito dalla call è: '%s'", err.Error())
		}
	case <-time.After(time.Second * time.Duration(utils.GetInt64EnvironmentVariable("MAX_WAITING_TIME_FOR_EDGE"))):
		return nil, fmt.Errorf("[*TIMEOUT_ERROR*] -> Non è stata ricevuta una risposta entro %d secondi da '%s'", utils.GetInt64EnvironmentVariable("MAX_WAITING_TIME_FOR_EDGE"), edgePeerPtr.PeerAddr)
	}

	return client, nil
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
	utils.PrintEvent("FILTER_NOTIFICATION_STARTED", "Inizio meccanismo di notifica dei filtri di bloom verso i vicini")
	FILTER_NOTIFY_TIME := utils.GetIntEnvironmentVariable("FILTER_NOTIFY_TIME")
	for {
		time.Sleep(time.Duration(FILTER_NOTIFY_TIME) * time.Second)
		err := notifyBloomFiltersToAdjacents()
		if err != nil {
			utils.PrintEvent("FILTER_ERROR", err.Error())
		}
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

	for edgePeer, adjConn := range adjacentsMap.peerConns {
		filterMessage := BloomFilterMessage{EdgePeer: SelfPeer, BloomFilter: filterEncode}
		call := adjConn.peerConnection.Go("EdgePeer.NotifyBloomFilter", filterMessage, new(int), nil)
		select {
		case <-call.Done:
			if call.Error != nil {
				return fmt.Errorf("impossibile inviare filtro di bloom a " + edgePeer.PeerAddr + ".\r\nL'errore restituito è: '" + call.Error.Error() + "'.")
			}
		case <-time.After(time.Second * time.Duration(utils.GetInt64EnvironmentVariable("MAX_WAITING_TIME_FOR_FILTER"))):
			return fmt.Errorf("[*TIMEOUT_ERROR*] -> Non è stata ricevuta una risposta entro %d secondi da '%s'", utils.GetInt64EnvironmentVariable("MAX_WAITING_TIME_FOR_FILTER"), edgePeer.PeerAddr)
		}

	}

	return nil
}

func NeighboursFileLookupWithoutFileRequest(fileName string, ttl int, requestID string, senderPeer string, forwarderPeer string, callbackServer string) {
	fileRequest := FileRequestMessage{
		FileName:       fileName,
		TTL:            ttl,
		RequestID:      requestID,
		SenderPeer:     EdgePeer{senderPeer},
		ForwarderPeer:  EdgePeer{forwarderPeer},
		CallbackServer: callbackServer}

	NeighboursFileLookup(fileRequest)
}

func NeighboursFileLookup(fileRequestMessage FileRequestMessage) {

	adjacentsMap.connsMutex.RLock()
	adjacentsMap.filtersMutex.RLock()

	defer adjacentsMap.filtersMutex.RUnlock()
	defer adjacentsMap.connsMutex.RUnlock()

	fileRequestMessage.ForwarderPeer = SelfPeer

	maxContactable := utils.GetIntEnvironmentVariable("MAX_CONTACTABLE_ADJ")
	contactedNum := 0

	// Contattiamo solo i vicini positivi ai filtri (tranne il mittente originario)
	for adj := range adjacentsMap.peerConns {
		adjFilter, isInMap := adjacentsMap.filterMap[adj]
		if isInMap {
			if adjFilter.Test([]byte(fileRequestMessage.FileName)) {
				contacted := contactNeighbourForFile(fileRequestMessage, adj)
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
			contacted := contactNeighbourForFile(fileRequestMessage, randomNeigh)
			if contacted {
				utils.PrintEvent("LOOKUP", "Richiesta inviata a "+randomNeigh.PeerAddr+" per complemento")
				contactedNum++
			}
			falseFiltersNeighbours = append(falseFiltersNeighbours[0:randomInt], falseFiltersNeighbours[randomInt+1:]...)
		}
	}
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

func contactNeighbourForFile(advancedFileRequestMessage FileRequestMessage, adj EdgePeer) bool {
	if adj != advancedFileRequestMessage.SenderPeer {
		adjConn := adjacentsMap.peerConns[adj]
		adjConn.peerConnection.Go("EdgePeer.FileLookup", advancedFileRequestMessage, new(int), nil)
		return true
	}
	return false
}
