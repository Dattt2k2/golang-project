package utils

import (
	"log"
	"os"
)

var (
	infoLogger  *log.Logger
	errorLogger *log.Logger
)

func init() {
	infoFile, err := os.OpenFile("info.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}

	errorFile, err := os.OpenFile("error.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}

	infoLogger = log.New(infoFile, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger = log.New(errorFile, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func Info(message string) {
	infoLogger.Println(message)
}

func Error(message string) {
	errorLogger.Println(message)
}