package cache

import (
	"edge/proto/file_transfer"
	"edge/redirection_channel"
	"edge/utils"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"sync"
	"syscall"
	"time"

	bloom "github.com/tylertreat/BoomFilters"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

/*
- Meccanismo di caching: manteniamo una coda di file nell'ordine in cui questi sono arrivati. Se arriva una richiesta (anche da un altro peer) per un file che
 abbiamo in cache, allora lo spostiamo di nuovo in fondo alla coda (come se fosse appena arrivato). In questo modo i file popolari non sono svantaggiati e
 difficilmente verrano buttati fuori dalla coda.
- Pochi file grandi oppure tanti piccoli in cache? -> Stabilire quando un file va salvato in cache tramite una threshold.
*/

type File struct {
	fileName   string
	fileSize   int64
	timestamp  time.Time
	popularity int
}

//implementazione sort.Interface per l'ordinamento dei file per popolarità crescente
type ByPopularity []File

func (f ByPopularity) Len() int           { return len(f) }
func (f ByPopularity) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
func (f ByPopularity) Less(i, j int) bool { return f[i].popularity < f[j].popularity }

type Cache struct {
	cachingList        []File
	cachingMap         map[string]File
	cacheMutex         sync.RWMutex
	fileInsertionMutex sync.RWMutex
}

var selfCache Cache = Cache{
	cachingList:        []File{},
	cachingMap:         map[string]File{},
	cacheMutex:         sync.RWMutex{},
	fileInsertionMutex: sync.RWMutex{},
}

func GetCache() *Cache {
	return &selfCache
}

func (cache *Cache) IsFileInCache(fileName string) bool {
	_, is := cache.cachingMap[fileName]
	return is
}

func (cache *Cache) GetFileSize(fileName string) int64 {

	for _, file := range cache.cachingList {
		if file.fileName == fileName {
			return file.fileSize
		}
	}

	return -1
}

//Inserisce un file all'interno della cache
func (cache *Cache) InsertFileInCache(redirectionChannel redirection_channel.RedirectionChannel, fileName string, fileSize int64) {
	// TODO gestire il problema per cui i file sono identificati soltanto dal nome (--> implementare login e accodare al nome del file quello dell'utenza)

	// Se esiste già, incrementa la popolarità del file
	if cache.incrementPopularity(fileName) {
		return
	} else {
		if IsFileCacheable(fileSize) {
			// Usiamo un diverso mutex per evitare che tutta la cache sia in blocco in caso di scrittura di un file. L'importante è che non ci siano scritture concorrenti
			cache.fileInsertionMutex.Lock()
			defer cache.fileInsertionMutex.Unlock()

			// SEZIONE CRITICA: soltanto un thread per volta può accedere a questa sezione altrimenti potrebbero esserci problemi durante l'eliminazione dei file
			err := cache.freeSpaceForInsert(fileSize)
			if err != nil {
				utils.PrintEvent("CACHE_SPACE_RELEASE_FAILURE", fmt.Sprintf("Errore durante la liberazione dello spazio in cache.\r\nL'errore è '%s'", err.Error()))
				return
			}
			err = writeChunksInCache(redirectionChannel, fileName)
			if err != nil {
				utils.PrintEvent("CACHE_WRITE_FAILURE", fmt.Sprintf("Errore durante la scrittura in cache.\r\nL'errore è '%s'", err.Error()))
				return
			}
			cache.insertFileInQueue(fileName, fileSize, 1)
			// FINE DELLA SEZIONE CRITICA-------------------------------------------------------------------------------------------------------------------------
		}
	}
}

// Incrementa la popolarità del file se questo esiste in cache. Ritorna true se ha avuto successo, false se il file non esiste.
func (cache *Cache) incrementPopularity(fileName string) bool {
	file, alreadyExists := cache.cachingMap[fileName]
	if alreadyExists {
		file.popularity++
		//TODO controllare se funziona, altrimenti aggiungere questa riga di codice:
		// cache.cachingMap[fileName] = file
		utils.PrintEvent("POPULARITY_INCREMENTED", fmt.Sprintf("La popolarità del File '%s' è salita al valore di '%d'", fileName, file.popularity))
		return true
	}
	return false
}

func (cache *Cache) insertFileInQueue(fileName string, fileSize int64, filePopularity int) {
	cache.cacheMutex.Lock()
	defer cache.cacheMutex.Unlock()

	newFile := File{fileName, fileSize, time.Now(), 1}
	// Inserimento del file nella mappa
	cache.cachingMap[fileName] = newFile
	// Inserimento del file nella coda
	cache.cachingList = append(cache.cachingList, newFile)
}

func (cache *Cache) RemoveFileFromCache(fileName string) {
	// Eliminazione del file dal filesystem
	err := removeWithLocks(fileName)
	if err != nil {
		utils.PrintEvent("CACHE_REMOVE_ERR", fmt.Sprintf("Errore durante l'eliminazione del file: '%s'", err.Error()))
		return
	}
	cache.removeFileFromQueue(fileName)
}

func (cache *Cache) GetFileForReading(fileName string) (*os.File, error) {
	localFile, err := os.Open(utils.GetEnvironmentVariable("FILES_PATH") + fileName)
	if err != nil {
		return nil, status.Error(codes.Code(file_transfer.ErrorCodes_FILE_NOT_FOUND_ERROR), fmt.Sprintf("[*ERROR*] - File opening failed.\r\nError: '%s'", err.Error()))
	}
	cache.incrementPopularity(fileName)
	return localFile, nil
}

func removeWithLocks(fileName string) error {
	// Apertura file
	filePath := utils.GetEnvironmentVariable("FILES_PATH") + fileName
	file, err := os.OpenFile(filePath, os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println("Errore durante l'apertura del file: ", err)
		return err
	}
	defer file.Close()
	// Tentiamo di prendere lock esclusivo sul file
	err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX)
	if err != nil {
		return fmt.Errorf("[*FILE_LOCK_ERR*] -> Errore nel tentativo di prendere Lock sul file '%s' ", fileName)
	}
	// Ora che abbiamo il lock, eliminiamo il file
	err = os.Remove(filePath)
	if err != nil {
		return fmt.Errorf("[*FILE_DELETE_ERR*] -> Errore nel tentativo di eliminare il Lock sul file '%s' ", fileName)
	}
	// Rilasciamo infine il lock
	err = syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
	if err != nil {
		return fmt.Errorf("[*FILE_UNLOCK_ERR*] -> Errore nel tentativo di rilasciare Lock sul file '%s' ", fileName)
	}
	return nil
}

func (cache *Cache) StartCache() {
	cache.ActivateCacheRecovery()
	go cache.cleanCachePeriodically()
}

func (cache *Cache) cleanCachePeriodically() {
	for {
		time.Sleep(time.Duration(utils.GetIntEnvironmentVariable("CACHE_CLEANING_FREQUENCY")))
		cache.deleteExpiredFiles()
	}
}

func (cache *Cache) deleteExpiredFiles() {
	for _, file := range cache.cachingList {
		if file.timestamp.Add(time.Duration(utils.GetInt64EnvironmentVariable("TIME_TO_DELETION"))).After(time.Now()) {
			cache.RemoveFileFromCache(file.fileName)
			utils.PrintEvent("FILE_DELETED", fmt.Sprintf("Il File '%s' è stato rimosso dalla cache a seguito della procedura periodica di cleanup.", file.fileName))
			convertAndPrintCachingMap(cache.cachingMap)
		}
	}
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
			cache.insertFileInQueue(file.Name(), file.Size(), 0)
		}
	}
	convertAndPrintCachingMap(cache.cachingMap)
}

func convertAndPrintCachingMap(cachingMap map[string]File) {
	printableCacheMap := make(map[string]byte)
	for fileName := range cachingMap {
		printableCacheMap[fileName] = 0
	}
	utils.PrintCustomMap(printableCacheMap, "Nessun file in cache...", "File trovati nella cache", "CACHE_RECOVERY_OK")
}

func (cache *Cache) removeFileFromQueue(file_name string) {
	cache.cacheMutex.Lock()
	defer cache.cacheMutex.Unlock()

	index, err := cache.getIndex(file_name)
	if err != nil {
		// La mappa e la coda sono inconsistenti, quindi attivare meccanismo di cache recovery
		cache.ActivateCacheRecovery()
		return
	}

	// Eliminazione file dalla mappa
	delete(cache.cachingMap, file_name)
	// Eliminazione file dalla coda
	cache.cachingList = append(cache.cachingList[:index], cache.cachingList[index+1:]...)
}

//Effettua un controllo sulla necessità di eliminare file dalla cache per liberare uno spazio pari a fileSize.
//In tal caso elimina file in fondo alla coda di popolarità fino a quando non si libera sufficiente memoria.
func (cache *Cache) freeSpaceForInsert(fileSize int64) error {
	cache.deleteExpiredFiles()
	freeMemory := retrieveFreeMemorySize()
	if freeMemory < fileSize {
		utils.PrintEvent("CACHING_QUEUE", fmt.Sprintln(cache.cachingList))
		sort.Sort(ByPopularity(cache.cachingList))
		utils.PrintEvent("SORTED_CACHING_QUEUE", fmt.Sprintln(cache.cachingList))
		return cache.chooseAndDeleteFiles(fileSize - freeMemory)
	}
	return nil
}

//Analizza i file necessari per liberare abbastanza spazio a partire da quelli a più bassa priorità.
//Controlla inoltre se tra i file scelti è possibile risparmiarne qualcuno minimizzando il numero di eliminazioni.
func (cache *Cache) chooseAndDeleteFiles(memoryRequired int64) error {
	var memoryCount int64 = 0
	var utilityFilesToDelete []File

	cache.cacheMutex.Lock()
	defer cache.cacheMutex.Unlock()
	// SEZIONE CRITICA: Accediamo ai file in cache per eliminarli -------------------------------------

	// Seleziona i file con popolarità più bassa
	for _, file := range cache.cachingList {
		memoryCount += file.fileSize
		utilityFilesToDelete = append(utilityFilesToDelete, file)
		if memoryCount >= memoryRequired {
			break
		}
	}

	// Se non è stato selezionato alcun file, allora c'è stato un errore.
	if len(utilityFilesToDelete) == 0 {
		return fmt.Errorf("impossibile individuare file da eliminare in cache")
	}

	var index int = len(utilityFilesToDelete) - 1
	var finalFilesToDelete []File

	// Controlla se è possibile evitare l'eliminazione di qualche file a partire da quelli con popolarità più alta
	for index > 0 {
		file := utilityFilesToDelete[index]
		// Se non eliminando il file non ho più abbastanza memoria, allora lo inserisco nei file da eliminare
		if memoryCount-file.fileSize < memoryRequired {
			finalFilesToDelete = append(finalFilesToDelete, file)
		}

		index--
	}

	// Eliminazione dei file
	for _, file := range finalFilesToDelete {
		cache.RemoveFileFromCache(file.fileName)
	}
	return nil
	// FINE DELLA SEZIONE CRITICA ---------------------------------------------------------------------
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

//Controlla se il file è adatto per essere inserito all'interno della cache.
//In particolare controlla se la dimensione è positiva e non superi un limite impostato nei parametri di configurazione.
func IsFileCacheable(file_size int64) bool {
	max_size := utils.GetInt64EnvironmentVariable("MAX_CACHABLE_FILE_SIZE")
	if file_size > 0 && file_size <= max_size {
		return true
	} else {
		if file_size <= 0 {
			utils.PrintEvent("CACHE_REFUSED", fmt.Sprintf("Il File potrebbe essere vuoto o non valido.\r\n(FILE_SIZE: %.2f MB)", float64(file_size)/1048576.0))
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
	for i, file := range cache.cachingList {
		if file.fileName == file_name {
			return i, nil
		}
	}
	return -1, errors.New("c'è inconsistenza tra la coda e la mappa della cache. Potrebbe essere dovuto ad accessi concorrenti non gestiti?")
}

// Calcola il filtro di bloom dei file in cache e lo restituisce.
func (cache *Cache) ComputeBloomFilter() *bloom.StableBloomFilter {
	cache.cacheMutex.RLock()
	defer cache.cacheMutex.RUnlock()

	newFilter := bloom.NewDefaultStableBloomFilter(
		utils.GetUintEnvironmentVariable("FILTER_N"),
		utils.GetFloatEnvironmentVariable("FALSE_POSITIVE_RATE"))
	for _, file := range cache.cachingList {
		newFilter.Add([]byte(file.fileName))
	}

	return newFilter
}

func writeChunksInCache(redirectionChannel redirection_channel.RedirectionChannel, fileName string) error {
	utils.PrintEvent("CACHE_WRITE_INIT", "La procedura di scrittura sulla cache è iniziata.")
	var localFile *os.File
	var fileCreated bool = false
	var err error = nil

	// Lettura dal canale di chunks
	for message := range redirectionChannel.MessageChannel {
		// TODO Capire se si può portare fuori
		if !fileCreated {
			localFile, err = os.Create(utils.GetEnvironmentVariable("FILES_PATH") + fileName)
			if err != nil {
				// Impossibile creare il file -> ritorniamo un errore
				utils.PrintEvent("CACHE_ERROR", "Creazione del file locale fallita")
				redirectionChannel.ReturnChannel <- err
				return err
			} else {
				// Il file è stato creato correttamente
				defer localFile.Close()
				fileCreated = true
				err = syscall.Flock(int(localFile.Fd()), syscall.F_WRLCK)
				if err != nil {
					utils.PrintEvent("FILE_LOCK_ERR", fmt.Sprintf("Errore nel tentativo di prendere Lock sul file '%s' ", fileName))
					redirectionChannel.ReturnChannel <- err
					return err
				}
				defer syscall.Flock(int(localFile.Fd()), syscall.F_UNLCK)
			}
		}

		// C'è stato un errore lato scrivente --> Rimozione file dalla cache
		if message.Err != nil {
			os.Remove(utils.GetEnvironmentVariable("FILES_PATH") + fileName)
			utils.PrintEvent("CACHE_ABORT", fmt.Sprintf("Notifica di errore ricevuta. Il file '%s' non verrà quindi caricato nella cache.\r\nErrore restituito: '%s'", fileName, err.Error()))
			redirectionChannel.ReturnChannel <- err
			return message.Err
		}

		// Scrittura file locale
		_, err = localFile.Write(message.Body)
		if err != nil {
			os.Remove(utils.GetEnvironmentVariable("FILES_PATH") + fileName)
			utils.PrintEvent("CACHE_FAILURE", fmt.Sprintf("Impossibile scrivere il file '%s' nella cache.\r\nErrore restituito: '%s'", fileName, err.Error()))
			redirectionChannel.ReturnChannel <- err
			return err
		}
	}

	if fileCreated {
		utils.PrintEvent("CACHE_SUCCESS", fmt.Sprintf("File '%s' caricato localmente con successo", fileName))
		redirectionChannel.ReturnChannel <- nil
		return nil
	}
	return fmt.Errorf("[*CACHE_ERR*] -> il file '%s' non è stato salvato in cache", fileName)
}
