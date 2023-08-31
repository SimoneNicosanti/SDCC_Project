package server

import (
	"edge/utils"
	"time"

	sortedmap "github.com/umpc/go-sortedmap"
	asc "github.com/umpc/go-sortedmap/asc"
)

/*
- Meccanismo di caching: manteniamo una coda di file nell'ordine in cui questi sono arrivati. Se arriva una richiesta (anche da un altro peer) per un file che
 abbiamo in cache, allora lo spostiamo di nuovo in fondo alla coda (come se fosse appena arrivato). In questo modo i file popolari non sono svantaggiati e
 difficilmente verrano buttati fuori dalla coda.
- Pochi file grandi oppure tanti piccoli in cache? -> Stabilire quando un file va salvato in cache.
*/

// Creazione di una Sorted Map con una size supposta di 50 valori e un ordinamento per tempo
var cachingMap *sortedmap.SortedMap

func init() {
	cachingMap = sortedmap.New(50, asc.Time)

}

func insertFileInQueue(file_name string, file_size int) {
	//TODO Controllare che ci sia abbastanza spazio e, in caso negativo, eliminare alcuni file

	if cachingMap.Has(file_name) {
		//Se esiste già, reinserisci il file in coda, ovvero eliminalo e reinseriscilo
		cachingMap.Delete(file_name)
	}

	cachingMap.Insert(file_name, time.Now())

}

func removeFileInQueue(file_name string) {
	//TODO Controllare che ci sia abbastanza spazio e, in caso negativo, eliminare alcuni file

	cachingMap.Delete(file_name)

}

func checkFileSize(file_size int) bool {
	if file_size > utils.GetIntegerEnvironmentVariable("MAX_CACHED_FILE_SIZE") {
		return false
	}
	return true
}
