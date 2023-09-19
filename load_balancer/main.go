package main

import "load_balancer/balancing"

func main() {
	balancing.ActAsBalancer()

	var forever chan struct{}
	<-forever
}
