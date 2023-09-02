package cache

import (
	"crypto/sha256"
	"edge/utils"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"syscall"

	bloom "github.com/tylertreat/BoomFilters"
)

/*
- Meccanismo di caching: manteniamo una coda di file nell'ordine in cui questi sono arrivati. Se arriva una richiesta (anche da un altro peer) per un file che
 abbiamo in cache, allora lo spostiamo di nuovo in fondo alla coda (come se fosse appena arrivato). In questo modo i file popolari non sono svantaggiati e
 difficilmente verrano buttati fuori dalla coda.
- Pochi file grandi oppure tanti piccoli in cache? -> Stabilire quando un file va salvato in cache tramite una threshold.
*/

type File struct {
	file_name string
	file_size int
}

type Cache struct {
	cachingQueue []File
	cachingMap   map[string]byte
	mutex        sync.RWMutex //TODO decidere chi prende i lock (se le singole funzioni o il chiamante)
}

var SelfCache Cache = Cache{
	cachingQueue: []File{},
	cachingMap:   map[string]byte{},
	mutex:        sync.RWMutex{}}

//TODO riportare tutto sulla cache vera

func (cache *Cache) InsertFileInCache(file_name string, file_size int) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	_, alreadyExists := cache.cachingMap[file_name]
	if alreadyExists {
		//Se esiste già, reinserisci il file in coda, ovvero eliminalo e reinseriscilo
		cache.RemoveFileInCache(file_name)
		cache.InsertFileInCache(file_name, file_size)
		return
	} else {
		if checkFileSize(file_size) {
			// Se non c'è abbastanza memoria disponibile elimino file fino a quando la memoria non basta
			//TODO Cambiare la gestione delle eliminazione (evitando di eliminare molti file)(?) -> Si potrebbero analizzare i file migliori da eliminare
			// piuttosto che eliminarli e basta
			for {
				//check sulla memoria disponibile
				if retrieveFreeMemorySize() < file_size {
					//elimino ultimo elemento nella queue
					//TODO eliminare anche il file -> altrimenti non si libererà mai lo spazio
					cache.RemoveFileInCache(cache.cachingQueue[len(cache.cachingQueue)-1].file_name)
				} else {
					break
				}
			}

			// Inserimento del file nella mappa
			cache.cachingMap[file_name] = 0
			// Inserimento del file nella coda
			cache.cachingQueue = append(cache.cachingQueue, File{file_name, file_size})
		}
	}
}

func (cache *Cache) RemoveFileInCache(file_name string) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	index, err := cache.getIndex(file_name)
	if err != nil {
		//TODO manage error
		utils.ExitOnError(err.Error(), err)
	}
	// Eliminazione file dalla mappa
	delete(cache.cachingMap, file_name)
	// Eliminazione file dalla coda
	cache.cachingQueue = append(cache.cachingQueue[:index], cache.cachingQueue[index+1:]...)
}

func retrieveFreeMemorySize() int {
	statPtr := new(syscall.Statfs_t)
	err := syscall.Statfs("/files", statPtr)
	if err != nil {
		//TODO manage error
		log.Println(err.Error())
	}
	return int(statPtr.Bfree * uint64(statPtr.Bsize))
}

func checkFileSize(file_size int) bool {
	return file_size <= utils.GetIntegerEnvironmentVariable("MAX_CACHED_FILE_SIZE")
}

/* Cerca l'indice dell'elemento da cercare. Ritorna l'indice se l'elemento è presente oppure -1 se non è presente.
Ritorna un errore nel caso in cui la mappa e la coda non sono consistenti nei valori contenuti. */
func (cache *Cache) getIndex(file_name string) (int, error) {
	_, alreadyExists := cache.cachingMap[file_name]
	if !alreadyExists {
		return -1, nil
	}
	for i, file := range cache.cachingQueue {
		if file.file_name == file_name {
			return i, nil
		}
	}
	return -1, errors.New("c'è inconsistenza tra la coda e la mappa della cache. Potrebbe essere dovuto ad accessi concorrenti non gestiti?")
}

func (cache *Cache) ComputeBloomFilter() *bloom.StableBloomFilter {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()

	newFilter := bloom.NewDefaultStableBloomFilter(
		utils.GetUintEnvironmentVariable("FILTER_N"),
		utils.GetFloatEnvironmentVariable("FALSE_POSITIVE_RATE"))
	for _, file := range cache.cachingQueue {
		newFilter.Add([]byte(file.file_name))
	}

	return newFilter
}

// TODO Potrebbero esserci dei problemi di blocco nel caso in cui il thread consumer vada in errore: il channel si satura e non si procede
func WriteChunksOnFile(fileChannel chan []byte, fileName string) error {

	localFile, err := os.Create("/files/" + fileName)
	if err != nil {
		return fmt.Errorf("[*ERROR*] - File creation failed")
	}
	syscall.Flock(int(localFile.Fd()), syscall.F_WRLCK)
	defer syscall.Flock(int(localFile.Fd()), syscall.F_UNLCK)
	defer localFile.Close()

	errorHashString := fmt.Sprintf("%x", sha256.Sum256([]byte("[*ERROR*]")))
	for chunk := range fileChannel {
		chunkString := fmt.Sprintf("%x", chunk)
		if strings.Compare(chunkString, errorHashString) == 0 {
			log.Println("[*ABORT*] -> Error occurred, removing file...")
			return os.Remove("/files/" + fileName)
		}
		_, err = localFile.Write(chunk)
		if err != nil {
			os.Remove("/files/" + fileName)
			log.Printf("[*ERROR*] -> Impossibile inserire il file '%s' in cache\n", fileName)
			return fmt.Errorf("[*ERROR*] -> Impossibile inserire il file in cache")
		}
	}
	log.Printf("[*SUCCESS*] - File '%s' caricato localmente con successo\r\n", fileName)

	return nil
}
