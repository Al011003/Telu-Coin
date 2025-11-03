package utils

import (
	"log"
	"os"
)

var (
	Log      *log.Logger
	MySQLLog *log.Logger
)

func init() {
	// General logger
	Log = log.New(os.Stdout, "[APP] ", log.Ldate|log.Ltime|log.Lshortfile)

	// MySQL specific logger
	MySQLLog = log.New(os.Stdout, "[MYSQL] ", log.Ldate|log.Ltime|log.Lshortfile)
}
