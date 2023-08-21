package main

import (
	"registry/services"
	"registry/utils"
)

func main() {

	utils.SetupEnvVariables("conf.properties")

	services.ActAsRegistry()

	var forever chan struct{}
	<-forever
}

// Looks for connected components inside network graph
