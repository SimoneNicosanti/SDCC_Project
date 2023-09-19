package balancing

import (
	"fmt"
	"load_balancer/proto"
	"log"
	"sync"
	"time"
)

type PeerServer struct {
	PeerAddr string
}

type HeartbeatMessage struct {
	PeerServer  PeerServer
	CurrentLoad int
}

type BalancingServer struct {
	proto.UnimplementedBalancingServiceServer
	mapMutex           sync.RWMutex
	peerServerMap      map[PeerServer]int
	heartbeatCheckTime time.Time
	heartbeats         map[PeerServer](time.Time)
}

func ExitOnError(errorMessage string, err error) {
	if err != nil {
		log.Println(errorMessage)
		log.Panicln(err.Error())
	}
}

func PrintEvent(title string, content string) {
	log.Printf("\033[1;30;47m[*" + title + "*]\033[0m")
	fmt.Printf(content + "\r\n\r\n")
}
