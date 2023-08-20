package peer

import (
	"edge/utils"
	"log"
	"net"
	"net/rpc"
	"strconv"
	"strings"
)

/*
	TODO : Logica di connessione ad altri Peer
	TODO : Logica di heartbeat (Peer -> Registry)
	TODO :
*/

type EdgePeer struct {
	PeerAddr string
}

//TODO : Risposta all'arrivo di un nuovo vicino

func ActAsPeer() {

	addresses, _ := net.InterfaceAddrs()
	ipAddr := strings.Split(addresses[1].String(), "/")[0]

	EdgePeerPtr := new(EdgePeer)
	edgeListenerPtr, errorMessage, err := registerServiceForEdge(ipAddr, EdgePeerPtr)
	utils.ExitOnError(errorMessage, err)

	registryClientPtr, errorMessage, err := registerToRegistry(EdgePeerPtr)
	utils.ExitOnError("Impossibile registrare il servizio sul registry server", err)

	//Notificare ai vicini il nuovo arrivato
	maxValue := len(Adjacent)
	for i = 0; i < maxValue; i++ {
		connect(Adjacent[i], ADJ_PORT)
	}

	//-------------------

	log.Println(edgeListenerPtr)
	defer registryClientPtr.Close()

}

func connect(addr string, port int) (*rpc.Client, string, error) {
	client, err := rpc.DialHTTP("tcp", addr+":"+strconv.Itoa(port))
	if err != nil {
		return nil, "Errore Dial HTTP", err
	}
	return client, "", err
}

func registerToRegistry(edgePeerPtr *EdgePeer) (*rpc.Client, string, error) {
	client, err := rpc.DialHTTP("tcp", "registry:1234")
	if err != nil {
		return nil, "Errore Dial HTTP", err
	}

	err = client.Call("RegistryService.PeerEnter", *edgePeerPtr, Adjacent)
	if err != nil {
		return nil, "Errore registrazione sul Registry Server", err
	}

	return client, "", nil
}

func registerServiceForEdge(ipAddrStr string, edgePeerPtr *EdgePeer) (*net.Listener, string, error) {
	err := rpc.Register(edgePeerPtr)
	if err != nil {
		return nil, "Errore registrazione del servizio", err
	}

	rpc.HandleHTTP()
	bindIpAddr := ipAddrStr + ":0"

	peerListener, err := net.Listen("tcp", bindIpAddr)
	if err != nil {
		return nil, "Errore listen", err
	}
	edgePeerPtr.PeerAddr = peerListener.Addr().String()

	return &peerListener, "", err
}

func (p *EdgePeer) Ping(input int, replyPtr *int) error {
	*replyPtr = 0
	return nil
}
