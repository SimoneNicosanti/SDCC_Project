package peer

import "log"

func (p *EdgePeer) Ping(none1 *int, none2 *int) error {
	log.Println("Ping ricevuto")
	return nil
}

func (p *EdgePeer) AddNeighbour(peer EdgePeer, none *int) error {
	_, err := connectAndAddNeighbour(peer)

	return err
}
