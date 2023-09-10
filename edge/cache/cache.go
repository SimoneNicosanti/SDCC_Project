package cache

import (
	"edge/channels"
	"edge/utils"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
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
	file_size int64
}

type Cache struct {
	cachingQueue []File
	cachingMap   map[string]byte
	mutex        sync.RWMutex
}

var selfCache Cache = Cache{
	cachingQueue: []File{},
	cachingMap:   map[string]byte{},
	mutex:        sync.RWMutex{}}

func GetCache() *Cache {
	return &selfCache
}

func (cache *Cache) IsFileInCache(file_name string) bool {
	_, is := cache.cachingMap[file_name]
	utils.PrintEvent("CACHING_MAP", fmt.Sprintln(cache.cachingMap))
	return is
}

func (cache *Cache) GetFileSize(file_name string) int64 {

	for _, file := range cache.cachingQueue {
		if file.file_name == file_name {
			return file.file_size
		}
	}

	return -1
}

func (cache *Cache) InsertFileInCache(redirectionChannel channels.RedirectionChannel, file_name string, file_size int64) {
	// TODO gestire il problema per cui i file sono identificati soltanto dal nome (versioning custom (?) // implementare login e fare un servizio per-user)
	// TODO Gestire il recupero dei file in cache

	_, alreadyExists := cache.cachingMap[file_name]
	if alreadyExists {
		// Se esiste già, reinserisci il file in testa alla coda
		cache.removeFileFromQueue(file_name)
		cache.insertFileInQueue(file_name, file_size)
		return
	} else {
		if IsFileCacheable(file_size) {
			freeMemoryForInsert(file_size, cache)
			err := writeChunksInCache(redirectionChannel, file_name)
			if err != nil {
				utils.PrintEvent("CACHE_WRITE_FAILURE", fmt.Sprintf("Errore durante la scrittura in cache.\r\nL'errore è '%s'", err.Error()))
				return
			}
			cache.insertFileInQueue(file_name, file_size)
		}
	}
}

func (cache *Cache) insertFileInQueue(file_name string, file_size int64) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	// Inserimento del file nella mappa
	cache.cachingMap[file_name] = 0
	// Inserimento del file nella coda
	cache.cachingQueue = append(cache.cachingQueue, File{file_name, file_size})
}

func (cache *Cache) RemoveFileFromCache(file_name string) {
	// Eliminazione del file dal filesystem//TODO prendere lock?
	err := os.Remove(utils.GetEnvironmentVariable("FILES_PATH") + file_name)
	if err != nil {
		utils.PrintEvent("CACHE_REMOVE_ERR", fmt.Sprintf("Errore durante l'eliminazione del file: '%s'", err.Error()))
		return
	}
	cache.removeFileFromQueue(file_name)
}

func (cache *Cache) ActivateCacheRecovery() {
	utils.PrintEvent("CACHE_RECOVERY_STARTED", "Trovata inconsistenza nella cache.\r\nProcedura di ripristino iniziata.")
	files, err := ioutil.ReadDir(utils.GetEnvironmentVariable("FILES_PATH"))
	if err != nil {
		utils.PrintEvent("CACHE_RECOVERY_ERROR", "Impossibile leggere i file nella cartella.")
		return
	}

	for _, file := range files {
		if !file.IsDir() {
			cache.insertFileInQueue(file.Name(), file.Size())
		}
	}
	utils.PrintCustomMap(cache.cachingMap, "Nessun file in cache...", "File trovati nella cache", "CACHE_RECOVERY_OK")
}

func (cache *Cache) removeFileFromQueue(file_name string) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	index, err := cache.getIndex(file_name)
	if err != nil {
		cache.ActivateCacheRecovery()
		return
	}

	// Eliminazione file dalla mappa
	delete(cache.cachingMap, file_name)
	// Eliminazione file dalla coda
	cache.cachingQueue = append(cache.cachingQueue[:index], cache.cachingQueue[index+1:]...)
}

func freeMemoryForInsert(file_size int64, cache *Cache) {
	// Se non c'è abbastanza memoria disponibile elimino file fino a quando la memoria non basta
	//TODO Cambiare la gestione delle eliminazione (evitando di eliminare molti file)(?) -> Si potrebbero analizzare i file migliori da eliminare
	// piuttosto che eliminare gli ultimi e basta
	for {
		if retrieveFreeMemorySize() < file_size {
			// lenght := len(cache.cachingQueue)
			// if lenght == 0 {
			// 	cache.RemoveFileFromCache(cache.cachingQueue[0].file_name)
			// }
			cache.RemoveFileFromCache(cache.cachingQueue[len(cache.cachingQueue)-1].file_name)
		} else {
			break
		}
	}
}

func retrieveFreeMemorySize() int64 {
	statPtr := new(syscall.Statfs_t)
	err := syscall.Statfs("/files", statPtr)
	if err != nil {
		utils.PrintEvent("RETR_FREE_MEM_ERR", fmt.Sprintf("Errore durante il recupero della dimensione della memoria libera.\r\nL'errore è '%s'", err.Error()))
	}
	//utils.PrintEvent("CACHE_INFO", fmt.Sprintf("Lo spazio residuo per lo storage di file è %d", statPtr.Bfree*uint64(statPtr.Bsize)))
	return int64(statPtr.Bfree * uint64(statPtr.Bsize))
}

func IsFileCacheable(file_size int64) bool {
	max_size := utils.GetInt64EnvironmentVariable("MAX_CACHABLE_FILE_SIZE")
	if file_size > 0 && file_size <= max_size {
		return true
	} else {
		if file_size <= 0 {
			utils.PrintEvent("CACHE_REFUSED", fmt.Sprintf("Il File potrebbe essere vuoto.\r\n(FILE_SIZE: %.2f MB)", float64(file_size)/1048576.0))
		} else {
			utils.PrintEvent("CACHE_REFUSED", fmt.Sprintf("Il File è troppo grande per essere caricato nella cache.\r\n(FILE_SIZE: %.2f MB > MAX: %.2f MB)", float64(file_size)/1048576.0, float64(max_size)/1048576.0))
		}
		return false
	}
}

// Cerca l'indice dell'elemento da cercare. Ritorna l'indice se l'elemento è presente oppure -1 se non è presente.
// Ritorna un errore nel caso in cui la mappa e la coda non sono consistenti nei valori contenuti.
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

func writeChunksInCache(redirectionChannel channels.RedirectionChannel, fileName string) error {
	utils.PrintEvent("CACHE_WRITE_INIT", "La procedura di scrittura sulla cache è terminata.")
	var localFile *os.File
	var fileCreated bool = false
	var errorOccurred bool = false
	var err error = nil
	var endLoop bool = false

	for !endLoop {
		select {
		// Lettura dal canale di chunks
		case chunk := <-redirectionChannel.ChunkChannel:
			if len(chunk) == 0 {
				endLoop = true
				break
			}
			if errorOccurred {
				break
			}
			// TODO Capire se si può portare fuori
			if !fileCreated {
				localFile, err = os.Create(utils.GetEnvironmentVariable("FILES_PATH") + fileName)
				if err != nil {
					// Impossibile creare il file -> consumiamo tutto sul canale e ritorniamo un errore
					errorOccurred = true
					utils.PrintEvent("CACHE_ERROR", "Creazione del file locale fallita")
					redirectionChannel.ReturnChannel <- err
				} else {
					// Creazione del file
					fileCreated = true
					err = syscall.Flock(int(localFile.Fd()), syscall.F_WRLCK)
					if err != nil {
						errorOccurred = true
						utils.PrintEvent("FILE_LOCK_ERR", fmt.Sprintf("Errore nel tentativo di prendere Lock sul file '%s' ", fileName))
						redirectionChannel.ReturnChannel <- err
					} else {
						defer syscall.Flock(int(localFile.Fd()), syscall.F_UNLCK)
					}
					defer localFile.Close()
				}
			}
			// Scrittura file locale
			_, err = localFile.Write(chunk)
			if err != nil {
				errorOccurred = true
				os.Remove(utils.GetEnvironmentVariable("FILES_PATH") + fileName)
				utils.PrintEvent("CACHE_FAILURE", fmt.Sprintf("Impossibile scrivere il file '%s' nella cache.\r\nErrore restituito: '%s'", fileName, err.Error()))
				redirectionChannel.ReturnChannel <- err
			}

		// Si è verificato un errore nello scrivente
		case err := <-redirectionChannel.ErrorChannel:
			if err != nil {
				os.Remove(utils.GetEnvironmentVariable("FILES_PATH") + fileName)
				utils.PrintEvent("CACHE_ABORT", fmt.Sprintf("Notifica di errore ricevuta. Il file '%s' non verrà quindi caricato nella cache.\r\nErrore restituito: '%s'", fileName, err.Error()))
				redirectionChannel.ReturnChannel <- err
			}
		}
	}
	if fileCreated && !errorOccurred {
		utils.PrintEvent("CACHE_SUCCESS", fmt.Sprintf("File '%s' caricato localmente con successo", fileName))
		redirectionChannel.ReturnChannel <- nil
		return nil
	}
	return err
}
