package peer

import (
	"edge/proto/client"
	"edge/utils"
	"fmt"
	"io"
	"log"
	"os"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	bloom "github.com/tylertreat/BoomFilters"
)

type EdgePeer struct {
	PeerAddr string
}

type PeerFileServer struct {
	client.UnimplementedEdgeFileServiceServer
	IpAddr string
}

// TODO Togliere il ping?? Ha sempre ragione il registry: potrebbe funzionare anche con, ma la logica rimane abbastanza simile
func (p *EdgePeer) Ping(edgePeer EdgePeer, returnPtr *int) error {
	log.Println("Ping Ricevuto da >>> ", edgePeer.PeerAddr)
	*returnPtr = 0
	return nil
}

func (p *EdgePeer) NotifyBloomFilter(bloomFilterMessage BloomFilterMessage, returnPtr *int) error {
	adjacentsMap.filtersMutex.Lock()
	defer adjacentsMap.filtersMutex.Unlock()
	edgePeer := bloomFilterMessage.EdgePeer
	edgeFilter := bloom.NewDefaultStableBloomFilter(
		utils.GetUintEnvironmentVariable("FILTER_N"),
		utils.GetFloatEnvironmentVariable("FALSE_POSITIVE_RATE"),
	)
	err := edgeFilter.GobDecode(bloomFilterMessage.BloomFilter)
	if err != nil {
		log.Println("[*ERROR*] -> impossibile decodificare il filtro ricevuto")
		return err
	}
	adjacentsMap.filterMap[edgePeer] = edgeFilter
	*returnPtr = 0
	log.Println("[*BLOOM_FILTER*] -> Ricevuto da " + edgePeer.PeerAddr)
	return nil
}

func (p *EdgePeer) AddNeighbour(peer EdgePeer, none *int) error {
	_, err := connectAndAddNeighbour(peer)

	return err
}

// TODO Aggiungere cache dei messaggi per non elaborare più volte lo stesso
func (p *EdgePeer) FileLookup(fileRequestMessage FileRequestMessage, returnPtr *PeerFileServer) error {
	_, err := os.Stat("/files/" + fileRequestMessage.FileName)
	fileRequestMessage.TTL--
	if os.IsNotExist(err) { //file NOT FOUND in local memory :/
		if fileRequestMessage.TTL > 0 {
			NeighboursFileLookup(fileRequestMessage)
		} else {
			// TTL <= 0 -> non propago la richiesta e non l'ho trovato --> fine corsa :')
			return fmt.Errorf("[*ERROR*] -> File '%s' wasn't found. Request TTL zeroed, not propagating request", fileRequestMessage.FileName)
		}
	} else if err == nil { //file FOUND in local memory --> i have it! ;)
		*returnPtr = peerFileServer
	} else { // Got an error :(
		return err
	}
	return nil
}

func (s *PeerFileServer) DownloadFromEdge(fileDownloadRequest *client.FileDownloadRequest, downloadStream client.EdgeFileService_DownloadFromEdgeServer) error {
	localFile, err := os.Open("/files/" + fileDownloadRequest.FileName)
	if err != nil {
		return status.Error(codes.Code(client.ErrorCodes_FILE_NOT_FOUND_ERROR), "[*ERROR*] -> File opening failed")
	}
	defer localFile.Close()
	chunkSize := utils.GetIntegerEnvironmentVariable("CHUNK_SIZE")
	buffer := make([]byte, chunkSize)
	for {
		n, err := localFile.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return status.Error(codes.Code(client.ErrorCodes_FILE_READ_ERROR), "[*ERROR*] -> Failed during read operation\r")
		}
		downloadStream.Send(&client.FileChunk{Chunk: buffer[:n]})
	}
	return nil
}
