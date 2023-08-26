package main

import (
	"edge/server"
	"edge/utils"
)

// "edge/server"

func main() {
	utils.SetupEnvVariables("conf.properties")
	// peer.ActAsPeer()
	server.ActAsServer()

	var forever chan struct{}
	<-forever

}
