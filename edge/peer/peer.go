package peer

import (
	"edge/utils"
	"log"
	"net"
	"net/rpc"
	"strings"
)

type EdgePeer struct {
	PeerId   int
	PeerAddr string
}

func ActAsPeer() {

	addresses, _ := net.InterfaceAddrs()
	ipAddr := strings.Split(addresses[1].String(), "/")[0]

	EdgePeerPtr := new(EdgePeer)
	edgeListenerPtr, err, errorMessage := registerServiceForEdge(ipAddr, EdgePeerPtr)
	utils.ExitOnError(errorMessage, err)

	registryClientPtr, err, errorMessage := registerToRegistry(EdgePeerPtr)
	utils.ExitOnError("Impossibile registrare il servizio sul registry server", err)

	log.Println(edgeListenerPtr)
	defer registryClientPtr.Close()

}

func registerToRegistry(edgePeerPtr *EdgePeer) (*rpc.Client, error, string) {
	client, err := rpc.DialHTTP("tcp", "registry:1234")
	if err != nil {
		return nil, err, "Errore Dial HTTP"
	}

	returnPtr := new(int)

	err = client.Call("RegistryService.PeerEnter", edgePeerPtr, returnPtr)
	if err != nil {
		return nil, err, "Errore registrazione sul Registry Server"
	}

	return client, nil, ""
}

func registerServiceForEdge(ipAddrStr string, edgePeerPtr *EdgePeer) (*net.Listener, error, string) {
	err := rpc.Register(edgePeerPtr)
	if err != nil {
		return nil, err, "Errore registrazione del servizio"
	}

	rpc.HandleHTTP()
	bindIpAddr := ipAddrStr + ":0"

	peerListener, err := net.Listen("tcp", bindIpAddr)
	if err != nil {
		return nil, err, "Errore listen"
	}
	edgePeerPtr.PeerAddr = peerListener.Addr().String()

	return &peerListener, nil, ""
}

func (p *EdgePeer) Ping(input int, replyPtr *int) error {
	*replyPtr = 0
	return nil
}
