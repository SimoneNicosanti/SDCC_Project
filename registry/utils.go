package main

import (
	"bufio"
	"errors"
	"log"
	"math/rand"
	"net/rpc"
	"os"
	"strconv"
	"strings"
	"sync"
)

type RegistryService int

type EdgePeer struct {
	PeerAddr string
}

type ConnectionMap struct {
	mutex       sync.RWMutex
	connections map[EdgePeer](*rpc.Client)
}

type GraphMap struct {
	mutex   sync.RWMutex
	peerMap map[EdgePeer]([]EdgePeer)
}

func ExitOnError(errorMessage string, err error) {
	if err != nil {
		log.Println(errorMessage)
		log.Panicln(err.Error())
	}
}

func SetupEnvVariables(fieldName string) {
	configMap, err := ReadConfigFile("conf.properties")
	ExitOnError("Impossibile leggere il file di configurazione", err)
	for key, value := range configMap {
		err := os.Setenv(key, value)
		ExitOnError("Impossibile impostare le variabili d'ambiente", err)
	}
}

func ReadConfigFile(filename string) (map[string]string, error) {
	config := make(map[string]string)

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			config[key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return config, nil
}

func GetEnvironmentVariable(variableName string) string {
	varibleString, isPresent := os.LookupEnv(variableName)
	if !isPresent {
		ExitOnError("Variabile d'ambiente non presente", errors.New("varibile non presente"))
	}
	return varibleString
}

func existsEdge() bool {
	randomNumber := rand.Float64()

	RAND_THR, err := strconv.ParseFloat(GetEnvironmentVariable("RAND_THR"), 64)
	ExitOnError("Impossibile fare il parsing di RAND_THR", err)
	return randomNumber > RAND_THR
}

func connectToPeer(edgePeer EdgePeer) (*rpc.Client, error) {
	client, err := rpc.DialHTTP("tcp", "registry:1234")
	if err != nil {
		return nil, err
	}
	return client, nil
}
