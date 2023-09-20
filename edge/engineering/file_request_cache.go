package engineering

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
	SenderPeer     string
	ForwarderPeer  string
}

// Attenzione alla gestione dei Mutex
type fileRequestCache struct {
	mutex      sync.RWMutex
	requestMap map[FileRequestMessage](time.Time)
}

var cache = fileRequestCache{
	sync.RWMutex{},
	map[FileRequestMessage]time.Time{},
}

func GetFileRequestCache() *fileRequestCache {
	return &cache
}

func (cache *fileRequestCache) AddRequestInCache(request FileRequestMessage) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	cache.requestMap[request] = time.Now()
}

//NB: Richiedere prima il lock del mutex
func (cache *fileRequestCache) removeRequestInCache(request FileRequestMessage) {
	delete(cache.requestMap, request)
}

func (cache *fileRequestCache) IsRequestAlreadyServed(fileRequestMessage FileRequestMessage) bool {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	var success bool = false
	var requestToDelete []FileRequestMessage = []FileRequestMessage{}

	for request, ttl := range cache.requestMap {
		var expirationTime time.Time = ttl.Add(time.Duration(utils.GetInt64EnvironmentVariable("REQUEST_CACHE_TTL")))
		if time.Now().After(expirationTime) {
			requestToDelete = append(requestToDelete, request)
		} else if request.RequestID == fileRequestMessage.RequestID && request.FileName == fileRequestMessage.FileName {
			success = true
		}
	}
	cache.requestMap[fileRequestMessage] = time.Now()

	// Eliminazione delle richieste scadute dalla cache (è improbabile che le riceveremo nuovamente)
	for _, request := range requestToDelete {
		delete(cache.requestMap, request)
	}

	return success
}
