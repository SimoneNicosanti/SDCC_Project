package utils

import (
	"bufio"
	"errors"
	"log"
	"os"
	"strings"
)

func ExitOnError(errorMessage string, err error) {
	if err != nil {
		log.Println(errorMessage)
		log.Panicln(err.Error())
	}
}

func GetConfigFieldFromFile(fileName string, fieldName string) (string, error) {
	config := make(map[string]string)

	file, err := os.Open(fileName)
	if err != nil {
		return "Impossibile aprire file di configurazione", err
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
		return "Impossibile leggere il file di configurazione", err
	}
	if value, exists := config[fieldName]; exists {
		return value, nil
	}

	err = errors.New("PROPERTY NOT FOUND ERROR")
	return "", err
}
