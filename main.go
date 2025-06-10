package main

import (
	"io"
	"log"
	"os"

	"github.com/pabotesu/kurohabaki-client/cmd"
)

func main() {
	logFile, err := os.OpenFile("/tmp/client.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("❌ Failed to open log file: %v", err)
	}
	multi := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multi)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	log.Println("kurohabaki client starting...")
	cmd.Execute()
}
