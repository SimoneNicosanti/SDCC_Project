package peer

import (
	"edge/utils"
	"fmt"
	"net/rpc"
	"time"
)

func pingsToAdjacents() {
	utils.PrintEvent("PING_STARTED", "Inizio meccanismo di ping verso i vicini")
	PING_FREQUENCY := utils.GetIntEnvironmentVariable("PING_FREQUENCY")
	for {
		time.Sleep(time.Duration(PING_FREQUENCY) * time.Second)
		pingFunction()
	}
}

func getConnections() {
	//TODO finire copia mappa dei vicini
	sourceSlice := adj...
	destinationSlice := make([]int, len(sourceSlice))

    	copy(destinationSlice, sourceSlice)
}

func pingFunction() {
	connections := getConnections()
	adjacentsMap.connsMutex.Lock()
	defer adjacentsMap.connsMutex.Unlock()

	for adjPeer, adjConn := range adjacentsMap.peerConns {
		call := adjConn.peerConnection.Go("EdgePeer.Ping", SelfPeer, new(int), nil)
		select {
		case <-call.Done:
			if call.Error != nil {
				timeoutAction(adjConn, adjPeer)
				break
			}
			// Se il peer risponde, allora azzero il numero di ping mancati
			adjConn.missedPing = 0
			adjacentsMap.peerConns[adjPeer] = adjConn
		case <-time.After(time.Second * time.Duration(utils.GetInt64EnvironmentVariable("MAX_WAITING_TIME_FOR_PING"))):
			// Il peer non risponde al Ping per X volte consecutive --> Rimozione dalla lista
			timeoutAction(adjConn, adjPeer)
		}
	}
}

func timeoutAction(adjConn AdjConnection, adjPeer EdgePeer) {
	adjConn.missedPing++
	adjacentsMap.peerConns[adjPeer] = adjConn
	utils.PrintEvent("PING_MISSED", fmt.Sprintf("Nessuna risposta da '%s' per la %d volta", adjPeer.PeerAddr, adjConn.missedPing))

	if adjConn.missedPing >= utils.GetIntEnvironmentVariable("MAX_MISSED_PING") {
		adjacentsMap.peerConns[adjPeer].peerConnection.Close()
		delete(adjacentsMap.peerConns, adjPeer)
		removeBloomFilter(adjPeer)
		utils.PrintEvent("NEIGHBOUR_REMOVED", fmt.Sprintf("Il vicino '%s' è stato rimosso dopo %d missed pings", adjPeer.PeerAddr, adjConn.missedPing))
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
	}
}

func heartbeatFunction() {
	heartbeatMessage := HeartbeatMessage{EdgePeer: SelfPeer}
	if registryClient == nil {
		newRegistryConnection, err := utils.ConnectToNode("registry:1234")
		if err != nil {
			utils.PrintEvent("REGISTRY_UNREACHABLE", "Impossibile stabilire connessione con il Registry")
			return
		}
		registryClient = newRegistryConnection
	}
	call := registryClient.Go("RegistryService.Heartbeat", heartbeatMessage, new(int), nil)
	select {
	case <-call.Done:
		if call.Error != nil {
			utils.PrintEvent("HEARTBEAT_ERROR", fmt.Sprintf("Invio di heartbeat al Registry fallito\r\nErrore: '%s'", call.Error.Error()))
			registryClient.Close()
			registryClient = nil
		}
	case <-time.After(time.Second * time.Duration(utils.GetInt64EnvironmentVariable("MAX_WAITING_TIME_FOR_EDGE"))):
		utils.PrintEvent("HEARTBEAT_ERROR", fmt.Sprintf("Timer scaduto.. Impossibile contattare il registry.\r\nErrore: '%s'", call.Error.Error()))
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
	client, err := utils.ConnectToNode(peer.PeerAddr)
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
