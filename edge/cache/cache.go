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

var selfCache Cache = Cache{
	cachingQueue: []File{},
	cachingMap:   map[string]byte{},
	mutex:        sync.RWMutex{}}

func GetCache() *Cache {
	return &selfCache
}

//TODO riportare tutto sulla cache vera

func (cache *Cache) InsertFileInCache(fileChannel chan []byte, file_name string, file_size int) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	cache.insertFileInCache(fileChannel, file_name, file_size, true)
}

func (cache *Cache) RemoveFileInCache(file_name string) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	cache.removeFileInCache(file_name, true)
}

func (cache *Cache) insertFileInCache(fileChannel chan []byte, file_name string, file_size int, writeFile bool) {
	//TODO gestire il problema per cui i file sono identificati soltanto dal nome

	_, alreadyExists := cache.cachingMap[file_name]
	if alreadyExists {
		// Se esiste già, reinserisci il file in coda, ovvero eliminalo e reinseriscilo
		cache.removeFileInCache(file_name, false)
		cache.insertFileInCache(fileChannel, file_name, file_size, false)
		return
	} else {
		if checkFileSize(file_size) {
			freeMemoryForInsert(file_size, cache)
			if writeFile {
				err := WriteChunksOnFile(fileChannel, file_name)
				if err != nil {
					//TODO gestione errori
					log.Println(err.Error())
					return
				}
			}
			// Inserimento del file nella mappa
			cache.cachingMap[file_name] = 0
			// Inserimento del file nella coda
			cache.cachingQueue = append(cache.cachingQueue, File{file_name, file_size})
		}
	}
}

func freeMemoryForInsert(file_size int, cache *Cache) {
	// Se non c'è abbastanza memoria disponibile elimino file fino a quando la memoria non basta
	//TODO Cambiare la gestione delle eliminazione (evitando di eliminare molti file)(?) -> Si potrebbero analizzare i file migliori da eliminare
	// piuttosto che eliminare gli ultimi e basta
	//check sulla memoria disponibile
	//elimino ultimo elemento nella queue
	for {
		if retrieveFreeMemorySize() < file_size {
			cache.removeFileInCache(cache.cachingQueue[len(cache.cachingQueue)-1].file_name, true)
		} else {
			break
		}
	}
}

func (cache *Cache) removeFileInCache(file_name string, removeFile bool) {

	index, err := cache.getIndex(file_name)
	if err != nil {
		//TODO manage error
		utils.ExitOnError(err.Error(), err)
	}

	// Eliminazione del file dal filesystem
	if removeFile {
		err = os.Remove("/files/" + file_name)
		if err != nil {
			fmt.Println("Errore durante l'eliminazione del file: ", err)
			return
		}
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
