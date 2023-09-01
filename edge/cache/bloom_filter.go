package cache

import (
	"edge/utils"
	"sync"

	bloom "github.com/tylertreat/BoomFilters"
)

type EdgeBloomFilter struct {
	Mutex   sync.RWMutex
	Changes int
	Filter  *bloom.CountingBloomFilter
}

var SelfBloomFilter EdgeBloomFilter

func SetupBloomFilterStruct() {
	SelfBloomFilter = EdgeBloomFilter{
		sync.RWMutex{},
		0,
		bloom.NewCountingBloomFilter(
			utils.GetUintEnvironmentVariable("FILTER_N"),
			utils.GetUint8EnvironmentVariable("BUCKET_NUMBER"),
			utils.GetFloatEnvironmentVariable("FALSE_POSITIVE_RATE"),
		),
	} // TODO Impostare parametri del filtro
}
