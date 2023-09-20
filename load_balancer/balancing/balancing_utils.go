package balancing

import (
	"bufio"
	"errors"
	"fmt"
	"load_balancer/proto"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type EdgeServer struct {
	ServerAddr string
}

type HeartbeatMessage struct {
	EdgeServer  EdgeServer
	CurrentLoad int
}

type BalancingServiceServer struct {
	proto.UnimplementedBalancingServiceServer
	mapMutex           sync.RWMutex
	edgeServerMap      map[EdgeServer]int
	heartbeatCheckTime time.Time
	heartbeats         map[EdgeServer](time.Time)
}

func ExitOnError(errorMessage string, err error) {
	if err != nil {
		log.Println(errorMessage)
		log.Panicln(err.Error())
	}
}

func PrintEvent(title string, content string) {
	log.Printf("\033[1;30;47m[*" + title + "*]\033[0m")
	fmt.Printf(content + "\r\n\r\n")
}

func SetupEnvVariables(fileName string) {
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
	variableString, isPresent := os.LookupEnv(variableName)
	if !isPresent {
		ExitOnError("Variabile d'ambiente non presente", errors.New("varibile non presente"))
	}
	return variableString
}

func GetIntEnvironmentVariable(variableName string) int {
	variableString := GetEnvironmentVariable(variableName)
	variableInt, err := strconv.ParseInt(variableString, 10, 64)
	ExitOnError("Impossibile convertire la variabile "+variableName, err)
	return int(variableInt)
}

func GetMyIPAddr() (string, error) {
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	// Loop through addresses to find a suitable one
	for _, addr := range addresses {
		ipNet, ok := addr.(*net.IPNet)
		if ok && !ipNet.IP.IsLoopback() {
			ipAddr := ipNet.IP.String()
			// Skip IPv6 addresses
			if ipNet.IP.To4() != nil {
				return ipAddr, nil
			}
		}
	}

	return "", fmt.Errorf("no suitable IP address found")
}
