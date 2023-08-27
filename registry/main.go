package main

import (
	"registry/services"
	"registry/utils"
)

// TODO Aggiungere limite al numero di vicini che ogni nodo può avere
func main() {

	utils.SetupEnvVariables("conf.properties")

	services.ActAsRegistry()

	var forever chan struct{}
	<-forever
}

// Looks for connected components inside network graph
