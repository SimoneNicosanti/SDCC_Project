package main

import (
	"edge/peer"
	"edge/server"
	"edge/utils"
)

// "edge/server"

func main() {
	utils.SetupEnvVariables("conf.properties")
	serverIpAddr := peer.ActAsPeer()
	server.ActAsServer(serverIpAddr)

	var forever chan struct{}
	<-forever

}
