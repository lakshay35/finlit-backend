package logging

import (
	"log"
	"os"
)

// WarningLogger ...
var WarningLogger *log.Logger

func init() {
	WarningLogger = log.New(os.Stdout, "[WARN] ", log.Ldate|log.Ltime|log.Lshortfile)
}
