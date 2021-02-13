package logging

import (
	"log"
	"os"
)

// InfoLogger ...
var InfoLogger *log.Logger

func init() {
	InfoLogger = log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime|log.Lshortfile)
}
