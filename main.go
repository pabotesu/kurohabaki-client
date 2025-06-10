package main

import (
	"log"
	"os"

	"github.com/pabotesu/kurohabaki-client/cmd"
	"golang.zx2c4.com/wireguard/device"
)

// カスタムWriterを定義（log.Writer に渡せるように）
type logWriter struct {
	write func(format string, args ...interface{})
}

func (lw logWriter) Write(p []byte) (n int, err error) {
	lw.write("%s", string(p))
	return len(p), nil
}

func main() {
	// ログファイルを開く
	logFile, err := os.OpenFile("/tmp/client.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("❌ Failed to open log file: %v", err)
	}
	defer logFile.Close()
	// WireGuard-Go の Logger を使って統合
	wgLogger := device.NewLogger(device.LogLevelVerbose, "kurohabaki")

	// log パッケージに WireGuard の Logger を流用
	log.SetOutput(logWriter{write: wgLogger.Verbosef})
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	log.Println("kurohabaki client starting...")
	cmd.Execute()
}
