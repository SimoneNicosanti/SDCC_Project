package peer

import (
	"edge/utils"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"strings"
	"time"
)

/*
	TODO : Logica di connessione ad altri Peer
	TODO : Azione nell'heartbeat (Peer -> Registry)
	TODO :
*/

type PeerService int

type EdgePeer struct {
	PeerAddr string
}

var selfPeer EdgePeer

func ActAsPeer() {
	addresses, _ := net.InterfaceAddrs()
	ipAddr := strings.Split(addresses[1].String(), "/")[0]

	edgePeerPtr := new(EdgePeer)
	edgeListenerPtr, errorMessage, err := registerServiceForEdge(ipAddr, edgePeerPtr)
	utils.ExitOnError(errorMessage, err)

	var adj []EdgePeer
	registryClientPtr, errorMessage, err := registerToRegistry(edgePeerPtr, &adj)
	utils.ExitOnError("Impossibile registrare il servizio sul registry server", err)

	startHeartBeatThread()

	//Connessione a tutti i vicini
	connectAndNotifyYourAdjacent(adj)

	//TODO creare un thread che risponde alle richieste degli altri peer
	//TODO thread che risponde alle richieste nella coda Rabbit_mq

	log.Println(edgeListenerPtr)
	defer registryClientPtr.Close()
}

func connectAndNotifyYourAdjacent(adj []EdgePeer) {
	for i := 0; i < len(adj); i++ {
		client, errorMessage, err := connectToPeer(adj[i].PeerAddr)
		utils.ExitOnError(errorMessage, err)
		peerConn := PeerConnection{adj[i], client}

		AddConnection(peerConn)

		err = CallAdjAddNeighbour(client, selfPeer)
		utils.ExitOnError(errorMessage, err)
	}
}

func connectToPeer(addr string) (*rpc.Client, string, error) {
	client, err := rpc.DialHTTP("tcp", addr)
	if err != nil {
		return nil, "Errore Dial HTTP", err
	}
	return client, "", err
}

func registerToRegistry(edgePeerPtr *EdgePeer, adj *[]EdgePeer) (*rpc.Client, string, error) {
	registryAddr := "registry:1234"
	client, err := rpc.DialHTTP("tcp", registryAddr)
	if err != nil {
		return nil, "Errore Dial HTTP", err
	}

	RegistryConn = PeerConnection{EdgePeer{registryAddr}, client}

	err = client.Call("RegistryService.PeerEnter", *edgePeerPtr, adj)
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

	selfPeer = EdgePeer{edgePeerPtr.PeerAddr}

	return &peerListener, "", err
}

func startHeartBeatThread() {
	hb := utils.GetConfigField("HEARTBEAT_FREQUENCY")
	heartbeatDuration, err := time.ParseDuration(hb + "s")
	if err != nil {
		fmt.Println("Error parsing heartbeat duration:", err)
		return
	}
	// Start the heartbeat thread
	go heartbeat(heartbeatDuration)
}

func heartbeat(interval time.Duration) {
	for {
		// TODO Perform heartbeat action here
		fmt.Println("Heartbeat action executed.")
		// Wait for the specified interval before the next heartbeat
		time.Sleep(interval)
	}
}

func (p *EdgePeer) AddNeighbour(peer EdgePeer, none *int) error {
	client, errorMessage, err := connectToPeer(peer.PeerAddr)
	utils.ExitOnError(errorMessage, err)
	peerConn := PeerConnection{peer, client}

	AddConnection(peerConn)

	return nil
}
