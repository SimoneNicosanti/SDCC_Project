package peer

import "log"

// TODO Togliere il ping?? Ha sempre ragione il registry: potrebbe funzionare anche con, ma la logica rimane abbastanza simile
func (p *EdgePeer) Ping(edgePeer EdgePeer, returnPtr *int) error {
	log.Println("Ping Ricevuto da >>> ", edgePeer.PeerAddr)
	*returnPtr = 0
	return nil
}

func (p *EdgePeer) AddNeighbour(peer EdgePeer, none *int) error {
	_, err := connectAndAddNeighbour(peer)

	return err
}

func (p *EdgePeer) GetFile(fileName string, returnPtr *int) error {
	return nil
}
