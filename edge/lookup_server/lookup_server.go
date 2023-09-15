package lookup_server

import (
	"bytes"
	"edge/peer"
	"edge/utils"
	"encoding/gob"
	"net"
)

type LookupServer struct {
	UdpAddr string
	conn    *net.UDPConn
}

func CreateLookupServer() *LookupServer {
	lookupServer := new(LookupServer)

	ipAddr, err := utils.GetMyIPAddr()
	if err != nil {
		utils.PrintEvent("LOOKUP_ERROR", "Errore nell'ottenimento dell'indirizzo")
		return nil
	}

	bindAddr := ipAddr + ":0"
	udpAddr, err := net.ResolveUDPAddr("udp", bindAddr)
	if err != nil {
		utils.PrintEvent("LOOKUP_ERROR", "Errore nella risoluzione dell'indirizzo")
		return nil
	}
	lookupServer.UdpAddr = udpAddr.String()

	// Crea una connessione UDP
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		utils.PrintEvent("LOOKUP_ERROR", "Errore nella creazione della connessione UDP")
		return nil
	}
	lookupServer.conn = conn

	return lookupServer
}

func ConnectToLookupServer(ipAddr string) *LookupServer {
	lookupServer := new(LookupServer)
	udpAddr, err := net.ResolveUDPAddr("udp", ipAddr)
	if err != nil {
		utils.PrintEvent("LOOKUP_ERROR", "Errore nella risoluzione dell'indirizzo")
		return nil
	}
	lookupServer.UdpAddr = udpAddr.String()

	// Crea una connessione UDP
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		utils.PrintEvent("LOOKUP_ERROR", "Errore nella creazione della connessione UDP")
		return nil
	}
	lookupServer.conn = conn

	return lookupServer
}

func (lookupServer *LookupServer) ReadFromServer(callbackChannel chan *peer.FileLookupResponse) {
	for {
		responsePtr, err := lookupServer.readFromChannel()
		if err == nil {
			callbackChannel <- responsePtr
		}
	}
}

func (lookupServer *LookupServer) readFromChannel() (*peer.FileLookupResponse, error) {
	// Leggi i dati dal datagramma UDP
	buffer := make([]byte, 128)
	n, _, err := lookupServer.conn.ReadFromUDP(buffer)
	if err != nil {
		utils.PrintEvent("LOOKUP_RCV_ERROR", "Errore durante la lettura dei dati UDP")
		return nil, err
	}

	gobData := buffer[:n]
	reader := bytes.NewReader(gobData)

	// Crea un decoder
	decoder := gob.NewDecoder(reader)
	lookupResponse := new(peer.FileLookupResponse)
	err = decoder.Decode(&lookupResponse)
	if err != nil {
		utils.PrintEvent("LOOKUP_DECODE_ERROR", "Errore durante la decodifica dei dati UDP")
		return nil, err
	}
	//fmt.Printf("Dati ricevuti da %s: %s\n", addr, datiRicevuti)
	return lookupResponse, nil
}

func (lookupServer *LookupServer) SendToServer(fileLookupResponse peer.FileLookupResponse) error {
	gobData := make([]byte, 128)
	writer := bytes.NewBuffer(gobData)

	gobEncoder := gob.NewEncoder(writer)
	err := gobEncoder.Encode(fileLookupResponse)
	if err != nil {
		utils.PrintEvent("LOOKUP_ENCODE_ERROR", "Errore durante la codifica dei dati UDP")
		return err
	}
	_, err = lookupServer.conn.Write(gobData)
	if err != nil {
		utils.PrintEvent("LOOKUP_SEND_ERROR", "Errore durante l'invio dei dati UDP")
		return err
	}

	return nil
}
