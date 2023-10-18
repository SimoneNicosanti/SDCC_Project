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

func cloneConnectionsMap() map[EdgePeer]AdjConnection {
	sourceMap := adjacentsMap.peerConns
	destinationMap := make(map[EdgePeer]AdjConnection, len(sourceMap))

	for key, value := range sourceMap {
		destinationMap[key] = value
	}

	return destinationMap
}

func pingFunction() {
	// Per evitare di bloccare la struttura delle connessioni troppo a lungo, ne facciamo una copia e lavoriamo su quella. Ogni volta che
	// si rende necessaria una modifica, allora prendiamo il lock
	adjacentsMap.connsMutex.Lock()
	connections := cloneConnectionsMap()
	adjacentsMap.connsMutex.Unlock()

	// NB: Stiamo lavorando su una copia della mappa, quindi le modifiche vanno fatte sulla struttura originale
	for adjPeer, adjConn := range connections {
		call := adjConn.peerConnection.Go("EdgePeer.Ping", SelfPeer, new(int), nil)
		select {
		case <-call.Done:
			if call.Error != nil {
				timeoutAction(adjPeer)
				break
			}
			// Se il peer risponde, allora azzero il numero di ping mancati
			// Prima di effettuare la modifica sulla struttura globale, prendiamo il lock
			adjacentsMap.connsMutex.Lock()
			connection, isStillInMap := adjacentsMap.peerConns[adjPeer]
			if isStillInMap {
				connection.missedPing = 0
				adjacentsMap.peerConns[adjPeer] = connection
			}
			adjacentsMap.connsMutex.Unlock()
		case <-time.After(time.Second * time.Duration(utils.GetInt64EnvironmentVariable("MAX_WAITING_TIME_FOR_PING"))):
			// Il peer non risponde al Ping per X volte consecutive --> Rimozione dalla lista
			timeoutAction(adjPeer)
		}
	}
}

func timeoutAction(adjPeer EdgePeer) {
	// Stiamo lavorando su una copia della mappa globale, quindi prima controlliamo che la connessione effettivamente esista.
	connection, isStillInMap := adjacentsMap.peerConns[adjPeer]
	if isStillInMap {
		// SEZIONE CRITICA ------------------------------------------------------------------------------
		adjacentsMap.connsMutex.Lock()
		defer adjacentsMap.connsMutex.Unlock()
		connection.missedPing++
		adjacentsMap.peerConns[adjPeer] = connection
		utils.PrintEvent("PING_MISSED", fmt.Sprintf("Nessuna risposta da '%s' per la %d volta", adjPeer.PeerAddr, connection.missedPing))

		if connection.missedPing >= utils.GetIntEnvironmentVariable("MAX_MISSED_PING") {
			adjacentsMap.peerConns[adjPeer].peerConnection.Close()
			delete(adjacentsMap.peerConns, adjPeer)
			removeBloomFilter(adjPeer)
			utils.PrintEvent("NEIGHBOUR_REMOVED", fmt.Sprintf("Il vicino '%s' è stato rimosso dopo %d missed pings", adjPeer.PeerAddr, connection.missedPing))
		}
		// FINE SEZIONE CRITICA -------------------------------------------------------------------------
	}

}

func removeBloomFilter(adjPeer EdgePeer) {
	// SEZIONE CRITICA ------------------------------------------------------------------------------
	adjacentsMap.filtersMutex.Lock()
	defer adjacentsMap.filtersMutex.Unlock()

	adjPeerConnection := adjacentsMap.peerConns[adjPeer]
	adjPeerConnection.missedPing = 0
	delete(adjacentsMap.filterMap, adjPeer)
	// FINE SEZIONE CRITICA -------------------------------------------------------------------------
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
