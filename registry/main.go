package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/rpc"
	"sync"
	"time"
)

var RAND_THR float64 = 0.6
var MONITOR_TIMER int = 60

type RegistryService int

type EdgePeer struct {
	PeerAddr string
}

type ConnectionMap struct {
	mutex       sync.RWMutex
	connections map[EdgePeer](*rpc.Client)
}

type GraphMap struct {
	mutex   sync.RWMutex
	peerMap map[EdgePeer]([]EdgePeer)
}

var connectionMap ConnectionMap = ConnectionMap{sync.RWMutex{}, make(map[EdgePeer](*rpc.Client))}

var graphMap GraphMap = GraphMap{sync.RWMutex{}, make(map[EdgePeer]([]EdgePeer))}

func ErrorFunction(errorMessage string, err error) {
	log.Println(errorMessage)
	log.Panicln(err.Error())
}

func main() {

	registryService := new(RegistryService)
	err := rpc.Register(registryService)
	if err != nil {
		ErrorFunction("Impossibile registrare il servizio", err)
	}

	rpc.HandleHTTP()
	list, err := net.Listen("tcp", ":1234")
	if err != nil {
		ErrorFunction("Impossibile mettersi in ascolto sulla porta", err)
	}

	fmt.Println("Waiting for connections...")

	go monitorNetwork()
	for {
		http.Serve(list, nil)
	}
}

func monitorNetwork() {
	for {
		time.Sleep(time.Duration(MONITOR_TIMER) * time.Second)
		saneMonitorPartitions()
	}
}

func saneMonitorPartitions() {
	graphMap.mutex.Lock()
	connectedComponents := findConnectedComponents()

	if len(connectedComponents) > 1 {
		unifyNetwork(connectedComponents)
	}
	graphMap.mutex.Unlock()
}

func unifyNetwork(connectedComponents [][]EdgePeer) {
	connectionMap.mutex.Lock()
	defer connectionMap.mutex.Unlock()
	for componentIndex := 0; componentIndex < len(connectedComponents)-1; componentIndex++ {
		firstComponent := connectedComponents[componentIndex]
		secondComponent := connectedComponents[componentIndex+1]
		unifyTwoComponents(firstComponent, secondComponent)
	}
}

func unifyTwoComponents(firstComponent []EdgePeer, secondComponent []EdgePeer) {
	for _, firstCompNode := range firstComponent {
		for _, secondCompNode := range secondComponent {
			if existsEdge() {
				graphMap.peerMap[firstCompNode] = append(graphMap.peerMap[firstCompNode], secondCompNode)
				graphMap.peerMap[secondCompNode] = append(graphMap.peerMap[secondCompNode], firstCompNode)

				//AddNeighbour: change data type
				// TODO Sistema il tipo di dato di PeerConnection
				firstNodeConn := connectionMap.connections[firstCompNode]
				secondNodeConn := connectionMap.connections[secondCompNode]

				err_1 := firstNodeConn.Call("PeerService.AddNeighbour", secondCompNode, nil)
				err_2 := secondNodeConn.Call("PeerService.AddNeighbour", firstCompNode, nil)
				if err_1 != nil {
					// TODO Manage error
				}
				if err_2 != nil {
					// TODO Manage error
				}

			}
		}
	}
}

func (r *RegistryService) PeerEnter(edgePeer EdgePeer, replyPtr *[]EdgePeer) error {

	connectionMap.mutex.Lock()
	defer connectionMap.mutex.Unlock()

	peerConnection, err := connectToPeer(edgePeer)
	if err != nil {
		return errors.New("Impossibile stabilire connessione con il peer")
	}
	connectionMap.mutex.Lock()
	connectionMap.connections[edgePeer] = peerConnection
	graphMap.mutex.Unlock()

	neighboursList := findNeighboursForPeer(edgePeer)
	graphMap.peerMap[edgePeer] = neighboursList

	for index := range neighboursList {
		neighbour := neighboursList[index]
		graphMap.peerMap[neighbour] = append(graphMap.peerMap[neighbour], edgePeer)
	}

	replyPtr = &neighboursList
	return nil
}

func connectToPeer(edgePeer EdgePeer) (*rpc.Client, error) {
	client, err := rpc.DialHTTP("tcp", "registry:1234")
	if err != nil {
		return nil, err
	}
	return client, nil
}

func findNeighboursForPeer(edgePeer EdgePeer) []EdgePeer {
	// peerNum := len(graphMap.peerMap)
	neighboursList := []EdgePeer{}
	for peer, _ := range graphMap.peerMap {
		if existsEdge() {
			neighboursList = append(neighboursList, peer)
		}
	}
	return neighboursList
}

func existsEdge() bool {
	randomNumber := rand.Float64()
	if randomNumber > RAND_THR {
		return true
	}
	return false
}

func (r *RegistryService) PeerExit(edgePeer EdgePeer, replyPtr *int) error {
	return nil
}

func (r *RegistryService) PeerPing(inputInt int, replyPtr *int) error {
	return nil
}
