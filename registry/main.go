package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
)

type RegistryService int

type EdgePeer struct {
	PeerId   int
	PeerAddr string
}

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

func (r *RegistryService) PeerEnter(edgePeer EdgePeer, replyPtr *int) error {
	fmt.Println(edgePeer.PeerAddr)
	fmt.Println(edgePeer.PeerId)
	*replyPtr = 0
	return nil
}

func (r *RegistryService) PeerExit(edgePeer EdgePeer, replyPtr *int) error {
	return nil
}

func (r *RegistryService) PeerPing(inputInt int, replyPtr *int) error {
	return nil
}
