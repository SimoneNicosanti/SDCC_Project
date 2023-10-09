package utils

import (
	"bufio"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math"
	"math/big"
	"net"
	"os"
	"strconv"
	"strings"
)

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
		ExitOnError("Variabile d'ambiente non presente", errors.New("variabile non presente"))
	}
	return variableString
}

func GetIntEnvironmentVariable(variableName string) int {
	variableString := GetEnvironmentVariable(variableName)
	variableInt, err := strconv.ParseInt(variableString, 10, 64)
	ExitOnError("Impossibile convertire la variabile "+variableName, err)
	return int(variableInt)
}

func GetFloatEnvironmentVariable(variableName string) float64 {
	variableFloat, err := strconv.ParseFloat(GetEnvironmentVariable(variableName), 64)
	ExitOnError("Impossibile convertire la variabile "+variableName, err)
	return variableFloat
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

func GenerateUniqueRandomID() int64 {
	max := big.NewInt(math.MaxInt64)

	// Generate a random integer in the range [0, max)
	randomInt, err := rand.Int(rand.Reader, max)
	if err != nil {
		return 0
	}
	randomInt64 := randomInt.Int64()
	return randomInt64
}

func ConvertToString(num int64) string {
	return strconv.FormatInt(num, 10)
}

func PrintCustomMap(customMap map[string]int, ifEmpty string, ifNotEmpty string, eventMessage string, hasValue bool) {
	listString := ""
	howMany := len(customMap)
	currentItemsNum := 0
	if howMany == 0 {
		listString = ifEmpty
	} else {
		listString = ifNotEmpty + ":\r\n"
		for item, value := range customMap {
			if hasValue {
				listString += "[*] " + item + " : " + fmt.Sprint(value)
			} else {
				listString += "[*] " + item
			}
			currentItemsNum++
			if currentItemsNum < howMany {
				listString += "\r\n"
			}
		}
	}
	PrintEvent(eventMessage, listString)
}
