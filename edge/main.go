package main

import (
	"edge/server"
	"edge/utils"
)

// "edge/server"

func main() {
	// peer.ActAsPeer()
	utils.SetupEnvVariables("conf.properties")
	server.ActAsServer()

}
