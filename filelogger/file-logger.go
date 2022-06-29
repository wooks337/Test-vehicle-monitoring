package filelogger

import (
	"log"
	"os"
)

var myLogger *log.Logger

func Init(filename string) {
	fpLog, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		myLogger = log.New(fpLog, "", log.LstdFlags)
	}
}
