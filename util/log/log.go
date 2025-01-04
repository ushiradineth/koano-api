package log

import (
	"log"
	"os"
)

type Logger struct {
	Info  *log.Logger
	Warn  *log.Logger
	Error *log.Logger
}

func New() *Logger {
	log := &Logger{
		Info:  log.New(os.Stdout, "[INFO]: ", log.Ldate|log.Ltime|log.Lshortfile),
		Warn:  log.New(os.Stdout, "[WARN]: ", log.Ldate|log.Ltime|log.Lshortfile),
		Error: log.New(os.Stdout, "[ERROR]: ", log.Ldate|log.Ltime|log.Lshortfile),
	}

	return log
}
