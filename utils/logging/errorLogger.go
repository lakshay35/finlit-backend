package logging

import (
	"log"
	"os"
)

// ErrorLogger ...
var ErrorLogger *log.Logger

func init() {
	ErrorLogger = log.New(os.Stdout, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile)
}
