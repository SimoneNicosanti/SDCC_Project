package peer

import (
	"edge/utils"
	"fmt"
	"log"
	"net/rpc"
	"time"
)

// TODO Aggiungere logica per cui se il Ping fallisce per x volte allora è dato per morto
// TODO Aggiungere logica per cui se ricevo Ping da un nodo che non è mio vicino allora lo aggiungo perché lo avevo dato per morto
func pingsToAdjacents() {
	utils.PrintEvent("PING_STARTED", "Inizio meccanismo di ping verso i vicini")
	PING_FREQUENCY := utils.GetIntEnvironmentVariable("PING_FREQUENCY")
	for {
		time.Sleep(time.Duration(PING_FREQUENCY) * time.Second)
		pingFunction()
	}
}

func pingFunction() {
	adjacentsMap.connsMutex.Lock()
	defer adjacentsMap.connsMutex.Unlock()

	for adjPeer, adjConn := range adjacentsMap.peerConns {
		err := adjConn.peerConnection.Call("EdgePeer.Ping", SelfPeer, new(int))
		if err != nil {
			adjConn.missedPing++
			adjacentsMap.peerConns[adjPeer] = adjConn
			utils.PrintEvent("MISSED_PING", fmt.Sprintf("Nessuna risposta da '%s' per la %d volta", adjPeer.PeerAddr, adjConn.missedPing))
			// Il peer non risponde al Ping per X volte consecutive --> Rimozione dalla lista
			if adjConn.missedPing >= utils.GetIntEnvironmentVariable("MAX_MISSED_PING") {
				adjacentsMap.peerConns[adjPeer].peerConnection.Close()
				delete(adjacentsMap.peerConns, adjPeer)
			}
		} else {
			// Se il peer risponde, allora azzero il numero di ping mancati
			removeBloomFilter(adjPeer)
		}
	}
}

func removeBloomFilter(adjPeer EdgePeer) {
	adjacentsMap.filtersMutex.Lock()
	defer adjacentsMap.filtersMutex.Unlock()

	adjPeerConnection := adjacentsMap.peerConns[adjPeer]
	adjPeerConnection.missedPing = 0
	delete(adjacentsMap.filterMap, adjPeer)
}

func Sprintf(s1, s2 string, i int) {
	panic("unimplemented")
}

func heartbeatToRegistry() {
	utils.PrintEvent("HEARTBEAT_STARTED", "Inizio meccanismo di heartbeat verso il server Registry")
	HEARTBEAT_FREQUENCY := utils.GetIntEnvironmentVariable("HEARTBEAT_FREQUENCY")
	for {
		time.Sleep(time.Duration(HEARTBEAT_FREQUENCY) * time.Second)
		heartbeatFunction()
		//utils.PrintEvent("HEARTBEAT", "heartbeat eseguito.")
	}
}

func heartbeatFunction() {
	heartbeatMessage := HeartbeatMessage{EdgePeer: SelfPeer}
	if registryClient == nil {
		newRegistryConnection, err := ConnectToNode("registry:1234")
		if err != nil {
			utils.PrintEvent("REGISTRY_UNREACHABLE", "Impossibile stabilire connessione con il Registry")
			return
		}
		registryClient = newRegistryConnection
	}
	err := registryClient.Call("RegistryService.Heartbeat", heartbeatMessage, new(int))
	if err != nil {
		utils.PrintEvent("HEARTBEAT_ERROR", "Invio di heartbeat al Registry fallito")
		log.Println(err.Error())
		registryClient.Close()
		registryClient = nil
	}

}

// func coerenceWithRegistry(registryAdjPeerList map[EdgePeer]byte) {
// 	// Aggiornare la lista dei peer in base a peer mancanti o meno:
// 	// - Ci sono tutti --> Perfetto

// 	for registryAdjPeer := range registryAdjPeerList {
// 		// - Manca qualcuno --> Lo aggiungo tra le connessioni
// 		_, isInMap := registryAdjPeerList[registryAdjPeer]
// 		if !isInMap {
// 			client, err := ConnectToNode(registryAdjPeer.PeerAddr)
// 			if err != nil {
// 				utils.PrintEvent("CONNECTION_ERROR", "Errore connessione a registry "+registryAdjPeer.PeerAddr)
// 			} else {
// 				adjacentsMap.peerConns[registryAdjPeer] = client
// 			}
// 		}
// 	}

// 	for adjacentNode := range adjacentsMap.peerConns {
// 		// - Qualcuno in più --> Lo rimuovo dalle connessioni
// 		_, isInMap := registryAdjPeerList[adjacentNode]
// 		if !isInMap {
// 			adjacentsMap.peerConns[adjacentNode].Close()
// 			delete(adjacentsMap.peerConns, adjacentNode)
// 		}
// 	}
// }

func connectAndAddNeighbour(peer EdgePeer) (*rpc.Client, error) {
	client, err := ConnectToNode(peer.PeerAddr)
	utils.PrintEvent("CONNECTION_ATTEMPT", "Tentativo di connessione a "+peer.PeerAddr)
	// Nel caso in cui uno dei vicini non rispondesse alla nostra richiesta di connessione,
	// il peer corrente lo ignorerà.
	if err != nil {
		utils.PrintEvent("CONNECTION_ERROR", "Impossibile stabilire la connessione con "+peer.PeerAddr)
		return nil, err
	}
	//Connessione con il vicino creata correttamente, quindi la aggiungiamo al nostro insieme di connessioni
	adjacentsMap.connsMutex.Lock()
	defer adjacentsMap.connsMutex.Unlock()
	adjacentsMap.peerConns[peer] = AdjConnection{client, 0}
	utils.PrintEvent("CONNECTION_SUCCESS", "Connessione con "+peer.PeerAddr+" effettuata con successo")
	return client, nil
}
