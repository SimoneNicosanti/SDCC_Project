package peer

import (
	"edge/cache"
	"edge/proto/client"
	"edge/utils"
	"fmt"
	"io"
	"os"
	"syscall"
	"time"

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
	client.UnimplementedEdgeFileServiceServer
	IpAddr string
}

func (p *EdgePeer) Ping(edgePeer EdgePeer, returnPtr *int) error {
	//utils.PrintEvent("PING_RECEIVED", "Ping Ricevuto da "+edgePeer.PeerAddr)
	*returnPtr = 0

	_, isPresent := adjacentsMap.peerConns[edgePeer]
	if !isPresent {
		conn, err := ConnectToNode(edgePeer.PeerAddr)
		if err != nil {
			utils.PrintEvent("PING_RCV_ERROR", fmt.Sprintf("Impossibile aprire connessione verso il nodo '%s'", edgePeer.PeerAddr))
			return err
		}
		adjacentsMap.connsMutex.Lock()
		adjacentsMap.peerConns[edgePeer] = AdjConnection{peerConnection: conn, missedPing: 0}
		adjacentsMap.connsMutex.Unlock()
		utils.PrintEvent("PING_STATUS", fmt.Sprintf("adjacentsMap.peerConns[edgePeer] : \r\n%s", fmt.Sprintln(adjacentsMap.peerConns[edgePeer])))
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

func (p *EdgePeer) FileLookup(fileRequestMessage FileRequestMessage, returnPtr *FileLookupResponse) error {
	utils.PrintEvent("LOOKUP_RECEIVED", "Richiesta ricevuta da "+fileRequestMessage.ForwarderPeer.PeerAddr)
	if checkServedRequest(fileRequestMessage) {
		utils.PrintEvent("LOOKUP_ABORT", "La richiesta relativa al ticket '"+fileRequestMessage.TicketId+"' è stata già servita.\r\nLa nuova richiesta verrà pertanto ignorata.")
		return fmt.Errorf("[*LOOKUP_ABORT*] -> Richiesta già servita")
	}
	fileRequestMessage.TTL--
	if !cache.GetCache().IsFileInCache(fileRequestMessage.FileName) { //file NOT FOUND in local memory :/
		if fileRequestMessage.TTL > 0 {
			utils.PrintEvent("LOOKUP_CONTINUE", fmt.Sprintf("Il File '%s' non è stato trovato in memoria. La richiesta viene inoltrata ad ulteriori vicini", fileRequestMessage.FileName))
			neighbourResponse, err := NeighboursFileLookup(fileRequestMessage)
			if err != nil {
				return err
			}
			*returnPtr = neighbourResponse
		} else {
			// TTL <= 0 -> non propago la richiesta e non l'ho trovato --> fine corsa :')
			return fmt.Errorf("[*LOOKUP_END*] -> Il File '%s' non è stato trovato. Il TTL della richiesta è pari a zero: la richiesta non verrà propagata", fileRequestMessage.FileName)
		}
	} else { //file FOUND in local memory --> i have it! ;)
		file_size := cache.GetCache().GetFileSize(fileRequestMessage.FileName)
		*returnPtr = FileLookupResponse{peerFileServer, file_size}
	}
	return nil
}

func checkServedRequest(fileRequestMessage FileRequestMessage) bool {
	fileRequestCache.mutex.Lock()
	defer fileRequestCache.mutex.Unlock()

	for message := range fileRequestCache.messageMap {
		if message.TicketId == fileRequestMessage.TicketId && message.FileName == fileRequestMessage.FileName && message.SenderPeer == fileRequestMessage.SenderPeer {
			return true
		}
	}
	fileRequestCache.messageMap[fileRequestMessage] = time.Now()
	return false

}

func (s *PeerFileServer) DownloadFromEdge(fileDownloadRequest *client.FileDownloadRequest, downloadStream client.EdgeFileService_DownloadFromEdgeServer) error {
	localFile, err := os.Open(utils.GetEnvironmentVariable("FILES_PATH") + fileDownloadRequest.FileName)
	if err != nil {
		return status.Error(codes.Code(client.ErrorCodes_FILE_NOT_FOUND_ERROR), "[*OPEN_ERROR*] -> Apertura del file fallita.")
	}
	//TODO Gestire errore FLOCK
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
			return status.Error(codes.Code(client.ErrorCodes_FILE_READ_ERROR), "[*READ_ERROR*] -> Faliure durante l'operazione di lettura.\r")
		}
		downloadStream.Send(&client.FileChunk{Chunk: buffer[:n]})
	}
	return nil
}
