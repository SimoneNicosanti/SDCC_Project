package utils

import "log"

func ExitOnError(errorMessage string, err error) {
	if err != nil {
		log.Println(errorMessage)
		log.Panicln(err.Error())
	}
}
