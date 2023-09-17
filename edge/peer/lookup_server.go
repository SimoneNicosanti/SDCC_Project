package peer

import (
	"bytes"
	"edge/utils"
	"encoding/gob"
	"fmt"
	"net"
)

type LookupServer struct {
	UdpAddr string
	conn    *net.UDPConn
	closed  bool
}

// Crea un server di callback per attendere risposte da chi trova il file nella cache
func CreateLookupServer() (*LookupServer, error) {
	lookupServer := new(LookupServer)

	ipAddr, err := utils.GetMyIPAddr()
	if err != nil {
		utils.PrintEvent("LOOKUP_ERROR", "Errore nell'ottenimento dell'indirizzo")
		return nil, err
	}

	udpAddr, err := net.ResolveUDPAddr("udp", ipAddr+":0")
	if err != nil {
		utils.PrintEvent("LOOKUP_ERROR", "Errore nella risoluzione dell'indirizzo")
		return nil, err
	}

	// Crea una connessione UDP
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		utils.PrintEvent("LOOKUP_ERROR", "Errore nella creazione della connessione UDP")
		return nil, err
	}
	lookupServer.UdpAddr = conn.LocalAddr().String()
	lookupServer.conn = conn
	lookupServer.closed = false
	return lookupServer, nil
}

func (lookupServer *LookupServer) CloseServer() {
	lookupServer.closed = true
	lookupServer.conn.Close()
}

// Permette di connettersi ad un server di callback lato client
func ConnectToLookupServer(ipAddr string) (*LookupServer, error) {
	lookupServer := new(LookupServer)
	udpAddr, err := net.ResolveUDPAddr("udp", ipAddr)
	if err != nil {
		utils.PrintEvent("LOOKUP_ERROR", "Errore nella risoluzione dell'indirizzo")
		return nil, err
	}
	lookupServer.UdpAddr = udpAddr.String()

	// Crea una connessione UDP
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		utils.PrintEvent("LOOKUP_ERROR", "Errore nella creazione della connessione UDP")
		return nil, err
	}
	lookupServer.conn = conn

	return lookupServer, nil
}

// Lettura risultati dal server di callback e di mandarli nel canale
func (lookupServer *LookupServer) ReadFromServer(callbackChannel chan *FileLookupResponse) {
	for {
		responsePtr, err := lookupServer.readFromChannel()
		if lookupServer.closed {
			close(callbackChannel)
			return
		}
		if err == nil {
			callbackChannel <- responsePtr
		}
	}
}

// Lettura e ritorno dei valori inviati al server
func (lookupServer *LookupServer) readFromChannel() (*FileLookupResponse, error) {
	// Leggi i dati dal datagramma UDP
	buffer := make([]byte, 128)
	n, _, err := lookupServer.conn.ReadFromUDP(buffer)
	utils.PrintEvent("RESPONSE_RECEIVED", fmt.Sprintf("Ricevuta un risposta dalla lookup"))
	if err != nil && !lookupServer.closed {
		utils.PrintEvent("LOOKUP_RCV_ERROR", fmt.Sprintf("Errore durante la lettura dei dati UDP: '%s'", err))
		return nil, err
	} else if err != nil && lookupServer.closed {
		return nil, nil
	}

	gobData := buffer[:n]
	reader := bytes.NewReader(gobData)

	// Crea un decoder
	decoder := gob.NewDecoder(reader)
	lookupResponse := new(FileLookupResponse)
	err = decoder.Decode(lookupResponse)
	if err != nil {
		utils.PrintEvent("LOOKUP_DECODE_ERROR", fmt.Sprintf("Errore durante la decodifica dei dati UDP: '%s'", err)) //TODO ******************
		return nil, err
	}
	//fmt.Printf("Dati ricevuti da %s: %s\n", addr, datiRicevuti)
	return lookupResponse, nil
}

// Invio di LookupResponse al server di callback contattato
func (lookupServer *LookupServer) SendToServer(fileLookupResponse FileLookupResponse) error {
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
