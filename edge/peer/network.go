package peer

import (
	"edge/utils"
	"errors"
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
			err := adjConn.Call("EdgePeer.Ping", selfPeer, new(int))
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
	// if err != nil {
	// 	log.Println("Error retreving HEARTBEAT_FREQUENCY from properties", err)
	// 	return
	// }

	for {
		// TODO Perform heartbeat action here
		// Wait for the specified interval before the next heartbeat
		time.Sleep(time.Duration(HEARTBEAT_FREQUENCY) * time.Second)
		adjacentsMap.connsMutex.Lock()

		heartbeatMessage := HeartbeatMessage{EdgePeer: selfPeer, NeighboursList: map[EdgePeer]byte{}}
		for adjPeer := range adjacentsMap.peerConns {
			heartbeatMessage.NeighboursList[adjPeer] = 0
		}
		returnMap := map[EdgePeer]byte{}
		if registryClient == nil {
			newRegistryConnection, err := ConnectToNode("registry:1234")
			if err != nil {
				log.Println("[*ERROR*] -> Impossibile stabilire connessione con il Registry")
				continue
			}
			registryClient = newRegistryConnection
		}
		err := registryClient.Call("RegistryService.Heartbeat", heartbeatMessage, &returnMap)
		if err != nil {
			log.Println("[*ERROR*] -> Failed to heartbeat to Registry")
			log.Println(err.Error())
			registryClient.Close()
			registryClient = nil
		}

		coerenceWithRegistry(returnMap)

		adjacentsMap.connsMutex.Unlock()
		//log.Println("[*HEARTBEAT*] -> executed.")
	}
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
				log.Println("[*ERROR*] -> Errore connessione " + registryAdjPeer.PeerAddr)
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
	//Nel caso in cui uno dei vicini non rispondesse alla nostra richiesta di connessione,
	// il peer corrente lo ignorerà.
	if err != nil {
		log.Println("[*ERROR*] -> Impossibile stabilire la connessione con " + peer.PeerAddr)
		return nil, errors.New("[*ERROR*] -> Impossibile stabilire la connessione con " + peer.PeerAddr)
	}

	//Connessione con il vicino creata correttamente, quindi la aggiungiamo al nostro insieme di connessioni
	adjacentsMap.connsMutex.Lock()
	adjacentsMap.peerConns[peer] = client
	defer adjacentsMap.connsMutex.Unlock()

	addConnection(peer, client)

	return client, nil
}
