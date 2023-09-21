package peer

import (
	"edge/cache"
	"edge/proto/file_transfer"
	"edge/utils"
	"fmt"
	"io"
	"os"
	"syscall"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	bloom "github.com/tylertreat/BoomFilters"
)

type EdgePeer struct {
	PeerAddr string
}

type FileLookupResponse struct {
	OwnerEdge PeerFileServer
	FileSize  int64
}

type PeerFileServer struct {
	file_transfer.UnimplementedEdgeFileServiceServer
	IpAddr string
}

func (p *EdgePeer) Ping(edgePeer EdgePeer, returnPtr *int) error {
	//utils.PrintEvent("PING_RECEIVED", "Ping Ricevuto da "+edgePeer.PeerAddr)
	*returnPtr = 0

	_, isPresent := adjacentsMap.peerConns[edgePeer]
	if !isPresent {
		conn, err := utils.ConnectToNode(edgePeer.PeerAddr)
		if err != nil {
			utils.PrintEvent("PING_RCV_ERROR", fmt.Sprintf("Impossibile aprire connessione verso il nodo '%s'", edgePeer.PeerAddr))
			return err
		}
		adjacentsMap.connsMutex.Lock()
		adjacentsMap.peerConns[edgePeer] = AdjConnection{peerConnection: conn, missedPing: 0}
		adjacentsMap.connsMutex.Unlock()
		utils.PrintEvent("NEIGHBOUR_RECOVERED", fmt.Sprintf("Il peer '%s' è stato riaggiunto tra i vicini.", edgePeer.PeerAddr))
	}

	return nil
}

func (p *EdgePeer) NotifyBloomFilter(bloomFilterMessage BloomFilterMessage, returnPtr *int) error {
	*returnPtr = 0
	edgePeer := bloomFilterMessage.EdgePeer
	edgeFilter := bloom.NewDefaultStableBloomFilter(
		utils.GetUintEnvironmentVariable("FILTER_N"),
		utils.GetFloatEnvironmentVariable("FALSE_POSITIVE_RATE"),
	)
	err := edgeFilter.GobDecode(bloomFilterMessage.BloomFilter)
	if err != nil {
		utils.PrintEvent("FILTER_ERROR", "impossibile decodificare il filtro di bloom ricevuto")
		return err
	}
	adjacentsMap.filtersMutex.Lock()
	defer adjacentsMap.filtersMutex.Unlock()
	adjacentsMap.filterMap[edgePeer] = edgeFilter
	//utils.PrintEvent("FILTER_RECEIVED", "filtro di bloom ricevuto correttamente da "+edgePeer.PeerAddr)
	return nil
}

func (p *EdgePeer) AddNeighbour(peer EdgePeer, none *int) error {
	_, err := connectAndAddNeighbour(peer)

	return err
}

func (p *EdgePeer) GetNeighbours(none int, returnPtr *map[EdgePeer]byte) error {
	adjacentsMap.connsMutex.RLock()
	defer adjacentsMap.connsMutex.RUnlock()
	returnAdjs := map[EdgePeer]byte{}
	for adjacent := range adjacentsMap.peerConns {
		returnAdjs[adjacent] = 0
	}
	*returnPtr = returnAdjs
	return nil
}

func (p *EdgePeer) FileLookup(fileRequestMessage FileRequestMessage, returnPtr *int) error {
	utils.PrintEvent("LOOKUP_RECEIVED", "Richiesta ricevuta da "+fileRequestMessage.ForwarderPeer.PeerAddr)
	//Se la richiesta è già stata servita una volta, la ignoriamo
	if GetFileRequestCache().IsRequestAlreadyServed(fileRequestMessage) {
		utils.PrintEvent("LOOKUP_ABORT", "La richiesta relativa al ticket '"+fileRequestMessage.RequestID+"' è stata già servita.\r\nLa nuova richiesta verrà pertanto ignorata.")
		return fmt.Errorf("[*LOOKUP_ABORT*] -> Richiesta già servita")
	}

	//Aggiungiamo la richiesta in cache e proviamo a soddisfarla
	GetFileRequestCache().AddRequestInCache(fileRequestMessage)

	fileRequestMessage.TTL--
	if !cache.GetCache().IsFileInCache(fileRequestMessage.FileName) { //file NOT FOUND in local memory :/
		if fileRequestMessage.TTL > 0 {
			utils.PrintEvent("LOOKUP_CONTINUE", fmt.Sprintf("Il File '%s' non è stato trovato in memoria. La richiesta viene inoltrata ad eventuali vicini", fileRequestMessage.FileName))
			NeighboursFileLookup(fileRequestMessage)
		} else {
			// TTL <= 0 -> non propago la richiesta e non l'ho trovato --> fine corsa :')
			utils.PrintEvent("LOOKUP_END", fmt.Sprintf("Il File '%s' non è stato trovato. Il TTL della richiesta è pari a zero: la richiesta non verrà propagata", fileRequestMessage.FileName))
			return fmt.Errorf("[*LOOKUP_END*] -> Il File richiesto non è stato trovato")
		}
	} else { //file FOUND in local memory --> i have it! ;)
		utils.PrintEvent("LOOKUP_HIT", fmt.Sprintf("Il File '%s' è stato trovato in memoria", fileRequestMessage.FileName))
		callbackServer, err := ConnectToLookupServer(fileRequestMessage.CallbackServer)
		if err != nil {
			utils.PrintEvent("LOOKUP_ERR", "Impossibile connettersi al server di callback")
			return err
		}
		defer callbackServer.CloseServer()

		file_size := cache.GetCache().GetFileSize(fileRequestMessage.FileName)
		lookupResponse := FileLookupResponse{peerFileServer, file_size}
		err = callbackServer.SendToServer(lookupResponse)
		if err != nil {
			utils.PrintEvent("LOOKUP_ERROR", "Errore invio risposta al server di callback")
			return err
		}
		utils.PrintEvent("LOOKUP_RESPONSE", fmt.Sprintf("Risposta per il File '%s' inviata", fileRequestMessage.FileName))

	}
	*returnPtr = 0
	return nil
}

func (s *PeerFileServer) DownloadFromEdge(fileDownloadRequest *file_transfer.FileDownloadRequest, downloadStream file_transfer.EdgeFileService_DownloadFromEdgeServer) error {
	localFile, err := os.Open(utils.GetEnvironmentVariable("FILES_PATH") + fileDownloadRequest.FileName)
	if err != nil {
		return status.Error(codes.Code(file_transfer.ErrorCodes_FILE_NOT_FOUND_ERROR), "[*OPEN_ERROR*] -> Apertura del file fallita.")
	}
	syscall.Flock(int(localFile.Fd()), syscall.F_RDLCK)
	defer syscall.Flock(int(localFile.Fd()), syscall.F_UNLCK)
	defer localFile.Close()

	chunkSize := utils.GetIntEnvironmentVariable("CHUNK_SIZE")
	buffer := make([]byte, chunkSize)
	for {
		n, err := localFile.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return status.Error(codes.Code(file_transfer.ErrorCodes_FILE_READ_ERROR), "[*READ_ERROR*] -> Faliure durante l'operazione di lettura.\r")
		}
		downloadStream.Send(&file_transfer.FileChunk{Chunk: buffer[:n]})
	}
	return nil
}
