package peer

import (
	"edge/utils"
	"encoding/json"
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
	if err != nil && !lookupServer.closed {
		utils.PrintEvent("LOOKUP_RCV_ERROR", fmt.Sprintf("Errore durante la lettura dei dati UDP: '%s'", err))
		return nil, err
	} else if err != nil && lookupServer.closed {
		return nil, nil
	}
	jsonData := buffer[:n]

	lookupResponsePtr := new(FileLookupResponse)
	err = json.Unmarshal(jsonData, lookupResponsePtr)
	if err != nil {
		utils.PrintEvent("LOOKUP_UNMARSHAL_ERROR", fmt.Sprintf("Errore durante l'unmarshal dei dati UDP: '%s'", err))
		return nil, err
	}
	return lookupResponsePtr, nil
}

// Invio di LookupResponse al server di callback contattato
func (lookupServer *LookupServer) SendToServer(fileLookupResponse FileLookupResponse) error {

	jsonData, err := json.Marshal(fileLookupResponse)
	if err != nil {
		utils.PrintEvent("LOOKUP_MARSHAL_ERROR", "Errore durante la codifica dei dati UDP")
		return err
	}

	_, err = lookupServer.conn.Write(jsonData)
	if err != nil {
		utils.PrintEvent("LOOKUP_SEND_ERROR", "Errore durante l'invio dei dati UDP")
		return err
	}

	return nil
}
