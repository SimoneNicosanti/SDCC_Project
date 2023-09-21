package main

import (
	"load_balancer/services"
	"load_balancer/utils"
)

func main() {

	utils.SetupEnvVariables("conf.properties")
	services.ActAsBalancer()

	var forever chan struct{}
	<-forever
}
