package peer

import (
	"edge/utils"
	"sync"
	"time"
)

type FileRequest struct {
	FileName      string
	RequestId     string
	ForwarderPeer EdgePeer
}

type FileLookupMessage struct {
	FileRequest
	TTL            int
	CallbackServer string
}

type FileDeleteMessage struct {
	FileRequest
}

// Attenzione alla gestione dei Mutex
type RequestCache struct {
	mutex      sync.RWMutex
	requestMap map[FileRequest](time.Time)
}

var requestCache = RequestCache{
	sync.RWMutex{},
	map[FileRequest]time.Time{},
}

func GetFileRequestCache() *RequestCache {
	return &requestCache
}

func (requestCache *RequestCache) AddRequestInCache(request FileRequest) {
	requestCache.mutex.Lock()
	defer requestCache.mutex.Unlock()

	requestCache.requestMap[request] = time.Now()
}

func (requestCache *RequestCache) IsRequestAlreadyServed(fileRequest FileRequest) bool {
	requestCache.mutex.Lock()
	defer requestCache.mutex.Unlock()

	var success bool = false
	var requestToDelete []FileRequest = []FileRequest{}

	for request, ttl := range requestCache.requestMap {
		var expirationTime time.Time = ttl.Add(time.Duration(utils.GetInt64EnvironmentVariable("REQUEST_CACHE_TTL")))
		if time.Now().After(expirationTime) {
			requestToDelete = append(requestToDelete, request)
		} else if request.RequestId == fileRequest.RequestId && request.FileName == fileRequest.FileName {
			success = true
		}
	}
	requestCache.requestMap[fileRequest] = time.Now()

	// Eliminazione delle richieste scadute dalla cache (è improbabile che le riceveremo nuovamente)
	for _, request := range requestToDelete {
		delete(requestCache.requestMap, request)
	}

	return success
}
