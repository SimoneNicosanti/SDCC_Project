package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/rpc"
	"sync"
)

var RAND_THR float64 = 0.6

type RegistryService int

type EdgePeer struct {
	PeerAddr string
}

type GraphMap struct {
	mutex   sync.RWMutex
	peerMap map[EdgePeer]([]EdgePeer)
}

var graphMap GraphMap = GraphMap{sync.RWMutex{}, make(map[EdgePeer]([]EdgePeer))}

func errorFunction(errorMessage string, err error) {
	log.Println(errorMessage)
	log.Panicln(err.Error())
}

func main() {

	registryService := new(RegistryService)
	err := rpc.Register(registryService)
	if err != nil {
		errorFunction("Impossibile registrare il servizio", err)
	}

	rpc.HandleHTTP()
	list, err := net.Listen("tcp", ":1234")
	if err != nil {
		errorFunction("Impossibile mettersi in ascolto sulla porta", err)
	}

	fmt.Println("Waiting for connections...")
	for {
		http.Serve(list, nil)
	}
}

func findConnectedComponents() {
	graphMap.mutex.Lock()
	defer graphMap.mutex.Unlock()

	visitedMap := make(map[EdgePeer](bool))
	for edge := range graphMap.peerMap {
		visitedMap[edge] = false
	}

	for peer := range graphMap.peerMap {
		foundedPeers := make([]EdgePeer, 10)
		recursiveConnectedComponentsResearch(peer, visitedMap, foundedPeers)
	}
}

func recursiveConnectedComponentsResearch(peer EdgePeer, visitedMap map[EdgePeer](bool), foundedPeers []EdgePeer) {

}

func (r *RegistryService) PeerEnter(edgePeer EdgePeer, replyPtr *[]EdgePeer) error {
	graphMap.mutex.Lock()
	defer graphMap.mutex.Unlock()

	neighboursList := findNeighboursForPeer(edgePeer)
	graphMap.peerMap[edgePeer] = neighboursList

	for index := range neighboursList {
		neighbour := neighboursList[index]
		graphMap.peerMap[neighbour] = append(graphMap.peerMap[neighbour], edgePeer)
	}

	replyPtr = &neighboursList
	return nil
}

func findNeighboursForPeer(edgePeer EdgePeer) []EdgePeer {
	// peerNum := len(graphMap.peerMap)
	neighboursList := []EdgePeer{}
	for peer, _ := range graphMap.peerMap {
		randomNumber := rand.Float64()
		if randomNumber > RAND_THR {
			neighboursList = append(neighboursList, peer)
		}
	}
	return neighboursList
}

func (r *RegistryService) PeerExit(edgePeer EdgePeer, replyPtr *int) error {
	return nil
}

func (r *RegistryService) PeerPing(inputInt int, replyPtr *int) error {
	return nil
}
