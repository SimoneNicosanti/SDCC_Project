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
)

/*
- Meccanismo di caching: manteniamo una coda di file nell'ordine in cui questi sono arrivati. Se arriva una richiesta (anche da un altro peer) per un file che
 abbiamo in cache, allora lo spostiamo di nuovo in fondo alla coda (come se fosse appena arrivato). In questo modo i file popolari non sono svantaggiati e
 difficilmente verrano buttati fuori dalla coda.
- Pochi file grandi oppure tanti piccoli in cache? -> Stabilire quando un file va salvato in cache tramite una threshold.
-
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

var cache Cache = Cache{cachingQueue: []File{}, cachingMap: map[string]byte{}, mutex: sync.RWMutex{}}

//TODO riportare tutto sulla cache vera
//TODO aggiornare filtri di Bloom ogni TOT operazioni

func (cache *Cache) InsertFileInQueue(file_name string, file_size int) {
	_, alreadyExists := cache.cachingMap[file_name]
	if alreadyExists {
		//Se esiste già, reinserisci il file in coda, ovvero eliminalo e reinseriscilo
		cache.RemoveFileInQueue(file_name)
		cache.InsertFileInQueue(file_name, file_size)
	} else {
		if checkFileSize(file_size) {
			//check sulla memoria disponibile
			if retrieveFreeMemorySize() < file_size {
				//elimino ultimo elemento nella queue
				cache.RemoveFileInQueue(cache.cachingQueue[len(cache.cachingQueue)-1].file_name)
			}
			cache.cachingMap[file_name] = 0
			cache.cachingQueue = append(cache.cachingQueue, File{file_name, file_size})
		}
	}
}

func (cache *Cache) RemoveFileInQueue(file_name string) {
	index, err := cache.getIndex(file_name)
	if err != nil {
		//TODO manage error

	}
	delete(cache.cachingMap, file_name)
	cache.cachingQueue = append(cache.cachingQueue[:index], cache.cachingQueue[index+1:]...)
}

func retrieveFreeMemorySize() int {
	statPtr := new(syscall.Statfs_t)
	err := syscall.Statfs("/files", statPtr)
	if err != nil {
		//TODO manage error
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

// TODO Potrebbero esserci dei problemi di blocco nel caso in cui il thread consumer vada in errore: il channel si satura e non si procede
// Potrebbe diventare un metodo di cache
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
