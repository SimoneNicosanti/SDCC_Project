package utils

import (
	"bufio"
	crypto_rand "crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
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

func SetupEnvVariables(fileName string) {
	configMap, err := ReadConfigFile(fileName)
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
		ExitOnError("Variabile d'ambiente non presente "+variableName, errors.New("variabile d'ambiente non presente"))
	}
	return variableString
}

func GetInt64EnvironmentVariable(variableName string) int64 {
	variableInt64, err := strconv.ParseInt(GetEnvironmentVariable(variableName), 10, 64)
	ExitOnError("Impossibile convertire la variabile "+variableName, err)
	return variableInt64
}

func GetIntEnvironmentVariable(variableName string) int {
	variableInt, err := strconv.ParseInt(GetEnvironmentVariable(variableName), 10, 64)
	ExitOnError("Impossibile convertire la variabile "+variableName, err)
	return int(variableInt)
}

func GetFloatEnvironmentVariable(variableName string) float64 {
	variableFloat, err := strconv.ParseFloat(GetEnvironmentVariable(variableName), 64)
	ExitOnError("Impossibile convertire la variabile "+variableName, err)
	return variableFloat
}

func GetUint8EnvironmentVariable(variableName string) uint8 {
	variableUint_8, err := strconv.ParseUint(GetEnvironmentVariable(variableName), 10, 8)
	ExitOnError("Impossibile convertire la variabile "+variableName, err)
	return uint8(variableUint_8)
}

func GetUintEnvironmentVariable(variableName string) uint {
	variableUint, err := strconv.ParseUint(GetEnvironmentVariable(variableName), 10, 32)
	ExitOnError("Impossibile convertire la variabile "+variableName, err)
	return uint(variableUint)
}

func isPortAvailable(port int) bool {
	listenAddr := fmt.Sprintf(":%d", port)
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return false
	}
	defer lis.Close()
	return true
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

func GenerateUniqueRandomID(existingIDs []string) (string, error) {
	for {
		randomBytes := make([]byte, 16)
		_, err := crypto_rand.Read(randomBytes)
		if err != nil {
			return "", err
		}
		randomID := base64.URLEncoding.EncodeToString(randomBytes)[:16]
		if !stringInSlice(randomID, existingIDs) {
			return randomID, nil
		}
	}

}

func stringInSlice(str string, slice []string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

func PrintEvent(title string, content string) {
	log.Printf("\033[1;30;47m[*" + title + "*]\033[0m")
	fmt.Printf(content + "\r\n\r\n")
}

func PrintCustomMap(customMap map[string]byte, ifEmpty string, ifNotEmpty string, eventMessage string) {
	listString := ""
	howMany := len(customMap)
	currentItemsNum := 0
	if howMany == 0 {
		listString = ifEmpty
	} else {
		listString = ifNotEmpty + ":\r\n"
		for item := range customMap {
			listString += "[*] " + item
			currentItemsNum++
			if currentItemsNum < howMany {
				listString += "\r\n"
			}
		}
	}
	PrintEvent(eventMessage, listString)
}
