package main

import (
	"github.com/pabotesu/kurohabaki-client/cmd"
	"github.com/pabotesu/kurohabaki-client/internal/logger"
)

func main() {
	logger.Println("kurohabaki client starting...")
	cmd.Execute()
}
