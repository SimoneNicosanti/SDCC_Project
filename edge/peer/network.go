package peer

import (
	"edge/utils"
	"log"
	"net/rpc"
	"time"
)

// TODO Aggiungere logica per cui se il Ping fallisce per x volte allora è dato per morto
// TODO Aggiungere logica per cui se ricevo Ping da un nodo che non è mio vicino allora lo aggiungo perché lo avevo dato per morto
func pingsToAdjacents() {
	PING_FREQUENCY := utils.GetIntegerEnvironmentVariable("PING_FREQUENCY")
	for {
		time.Sleep(time.Duration(PING_FREQUENCY) * time.Second)
		adjacentsMap.connsMutex.Lock()
		for adjPeer, adjConn := range adjacentsMap.peerConns {
			err := adjConn.Call("EdgePeer.Ping", SelfPeer, new(int))
			if err != nil {
				// Il peer non risponde al Ping --> Rimozione dalla lista
				adjacentsMap.peerConns[adjPeer].Close()
				delete(adjacentsMap.peerConns, adjPeer)
			}
		}
		adjacentsMap.connsMutex.Unlock()
	}
}

func heartbeatToRegistry() {
	HEARTBEAT_FREQUENCY := utils.GetIntegerEnvironmentVariable("HEARTBEAT_FREQUENCY")
	for {
		time.Sleep(time.Duration(HEARTBEAT_FREQUENCY) * time.Second)
		heartbeatFunction()
		//utils.PrintEvent("HEARTBEAT", "heartbeat eseguito.")
	}
}

func heartbeatFunction() {
	adjacentsMap.connsMutex.Lock()
	defer adjacentsMap.connsMutex.Unlock()

	heartbeatMessage := HeartbeatMessage{EdgePeer: SelfPeer, NeighboursList: map[EdgePeer]byte{}}
	for adjPeer := range adjacentsMap.peerConns {
		heartbeatMessage.NeighboursList[adjPeer] = 0
	}
	returnMap := map[EdgePeer]byte{}
	if registryClient == nil {
		newRegistryConnection, err := ConnectToNode("registry:1234")
		if err != nil {
			utils.PrintEvent("CONNECTION_ERROR", "Impossibile stabilire connessione con il Registry")
			return
		}
		registryClient = newRegistryConnection
	}
	err := registryClient.Call("RegistryService.Heartbeat", heartbeatMessage, &returnMap)
	if err != nil {
		utils.PrintEvent("HEARTBEAT_ERROR", "Invio di heartbeat al Registry fallito")
		log.Println(err.Error())
		registryClient.Close()
		registryClient = nil
	}

	coerenceWithRegistry(returnMap)

}

func coerenceWithRegistry(registryAdjPeerList map[EdgePeer]byte) {
	// Aggiornare la lista dei peer in base a peer mancanti o meno:
	// - Ci sono tutti --> Perfetto

	for registryAdjPeer := range registryAdjPeerList {
		// - Manca qualcuno --> Lo aggiungo tra le connessioni
		_, isInMap := registryAdjPeerList[registryAdjPeer]
		if !isInMap {
			client, err := ConnectToNode(registryAdjPeer.PeerAddr)
			if err != nil {
				utils.PrintEvent("CONNECTION_ERROR", "Errore connessione a registry "+registryAdjPeer.PeerAddr)
			} else {
				adjacentsMap.peerConns[registryAdjPeer] = client
			}
		}
	}

	for adjacentNode := range adjacentsMap.peerConns {
		// - Qualcuno in più --> Lo rimuovo dalle connessioni
		_, isInMap := registryAdjPeerList[adjacentNode]
		if !isInMap {
			adjacentsMap.peerConns[adjacentNode].Close()
			delete(adjacentsMap.peerConns, adjacentNode)
		}
	}
}

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
	adjacentsMap.peerConns[peer] = client
	utils.PrintEvent("CONNECTION_SUCCESS", "Connessione con "+peer.PeerAddr+" effettuata con successo")
	return client, nil
}
