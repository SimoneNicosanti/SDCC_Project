package peer

import (
	"edge/utils"
	"sync"
	"time"
)

type FileRequestMessage struct {
	FileName       string
	TTL            int
	RequestID      string
	CallbackServer string
	SenderPeer     EdgePeer
	ForwarderPeer  EdgePeer
}

// Attenzione alla gestione dei Mutex
type RequestCache struct {
	mutex      sync.RWMutex
	requestMap map[FileRequestMessage](time.Time)
}

var requestCache = RequestCache{
	sync.RWMutex{},
	map[FileRequestMessage]time.Time{},
}

func GetFileRequestCache() *RequestCache {
	return &requestCache
}

func (requestCache *RequestCache) AddRequestInCache(request FileRequestMessage) {
	requestCache.mutex.Lock()
	defer requestCache.mutex.Unlock()

	requestCache.requestMap[request] = time.Now()
}

func (requestCache *RequestCache) IsRequestAlreadyServed(fileRequestMessage FileRequestMessage) bool {
	requestCache.mutex.Lock()
	defer requestCache.mutex.Unlock()

	var success bool = false
	var requestToDelete []FileRequestMessage = []FileRequestMessage{}

	for request, ttl := range requestCache.requestMap {
		var expirationTime time.Time = ttl.Add(time.Duration(utils.GetInt64EnvironmentVariable("REQUEST_CACHE_TTL")))
		if time.Now().After(expirationTime) {
			requestToDelete = append(requestToDelete, request)
		} else if request.RequestID == fileRequestMessage.RequestID && request.FileName == fileRequestMessage.FileName {
			success = true
		}
	}
	requestCache.requestMap[fileRequestMessage] = time.Now()

	// Eliminazione delle richieste scadute dalla cache (è improbabile che le riceveremo nuovamente)
	for _, request := range requestToDelete {
		delete(requestCache.requestMap, request)
	}

	return success
}
