package log

import (
	"log"
	"os"
)

var Logger *log.Logger = log.New(os.Stdout, "j2z: ", 0)
