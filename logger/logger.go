package logger

import (
	"log"
	"os"
)

var logger *log.Logger

func Init() *log.Logger {
	logger = log.New(os.Stderr, "", log.LstdFlags|log.Llongfile)
	return logger
}

func GetLogger() log.Logger {
	return *logger
}
