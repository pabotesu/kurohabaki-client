package cmd

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/pabotesu/kurohabaki-client/config"
	"github.com/pabotesu/kurohabaki-client/internal/agent"
	"github.com/pabotesu/kurohabaki-client/internal/etcd"
	"github.com/pabotesu/kurohabaki-client/internal/logger"
	"github.com/pabotesu/kurohabaki-client/internal/util"
	"github.com/pabotesu/kurohabaki-client/internal/wg"
	"github.com/spf13/cobra"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

var (
	configPath string
	debugMode  bool // Debug flag specific to up command
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Start WireGuard interface and connect to peers",
	PreRun: func(cmd *cobra.Command, args []string) {
		// Initialize logger with debug mode setting
		logger.Init(debugMode)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if agent is already running
		pidFile := util.GetPidFilePath()
		if _, err := os.Stat(pidFile); err == nil {
			return fmt.Errorf("agent is already running. Use 'down' command to stop it first")
		}

		logger.Println("Bringing up WireGuard interface...")

		cfg, err := config.Load(configPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Use fixed interface name for now
		ifaceName := "kh0"
		wgIf, err := wg.NewWireGuardInterface(ifaceName)
		if err != nil {
			return fmt.Errorf("failed to create interface: %w", err)
		}

		if err := wgIf.AddAddress(cfg.Interface.Address); err != nil {
			return fmt.Errorf("failed to add address: %w", err)
		}

		if err := wgIf.SetUpInterface(); err != nil {
			return fmt.Errorf("failed to set interface up: %w", err)
		}

		conf := wg.BuildWGConfig(cfg)
		if err := wgIf.Up(conf); err != nil {
			return fmt.Errorf("failed to apply WireGuard config: %w", err)
		}
		logger.Println("WireGuard interface is up")
		// Prevent process from exiting to keep interface alive

		// Configure etcd logging based on debug mode
		etcd.ConfigureEtcdLogger(debugMode)

		// Get the logger
		zapLogger := zap.L()

		// Initialize etcd client with custom logger
		etcdCli, err := clientv3.New(clientv3.Config{
			Endpoints:   []string{cfg.Etcd.Endpoint},
			DialTimeout: 5 * time.Second,
			Logger:      zapLogger,
		})
		if err != nil {
			return fmt.Errorf("failed to connect to etcd: %w", err)
		}

		// IMPORTANT: Only close etcdClient in parent process or debug mode
		if os.Getenv("KH_BACKGROUND") != "1" {
			defer etcdCli.Close()
		}

		// Check etcd health
		if err := etcd.CheckEtcdHealth(etcdCli); err != nil {
			// 改行を避け、一貫した形式でログを出力
			logger.Println("⚠️ Warning: " + err.Error())
			logger.Println("⚠️ Will continue with local configuration but peer discovery may not work")
			// Don't return error here, allow to continue with local config
		}

		privKey, err := wgtypes.ParseKey(cfg.Interface.PrivateKey)
		if err != nil {
			return fmt.Errorf("failed to parse private key: %w", err)
		}
		pubKey := privKey.PublicKey()
		selfPubKey := base64.StdEncoding.EncodeToString(pubKey[:])
		logger.Printf("🔑 selfPubKey: %s", selfPubKey)
		logger.Printf("🔎 Peer count in conf: %d", len(conf.Peers))
		logger.Printf("✅ Peers in config: %d", len(conf.Peers))
		logger.Printf("✅ PublicKey (self): %s", selfPubKey)
		logger.Printf("✅ etcd endpoint: %s", cfg.Etcd.Endpoint)
		logger.Println("✅ Starting Agent...")

		// Debug mode behavior differs from normal mode
		if debugMode {
			// In debug mode, run in foreground with signals
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Handle signals for graceful shutdown
			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

			go func() {
				sig := <-sigCh
				logger.Printf("🛑 Caught signal: %v, shutting down...", sig)
				cancel()
			}()

			a := agent.New(wgIf, etcdCli, selfPubKey)
			a.Run(ctx)

			logger.Println("🏁 Agent stopped, exiting normally.")
			return nil
		} else {
			// Non-debug mode: run in background

			// Fork a child process that will continue running
			if os.Getenv("KH_BACKGROUND") != "1" {
				// Parent process - fork and exit
				cmd := exec.Command(os.Args[0], append([]string{"up", "--config", configPath}, os.Args[2:]...)...)
				cmd.Env = append(os.Environ(), "KH_BACKGROUND=1")

				// Redirect stdout and stderr to log file
				logFilePath := "/var/log/kh-client.log"
				logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					return fmt.Errorf("failed to open log file: %w", err)
				}
				defer logFile.Close()

				cmd.Stdout = logFile
				cmd.Stderr = logFile

				// Create a new process group so signals to the parent don't affect the child
				cmd.SysProcAttr = &syscall.SysProcAttr{
					Setsid: true, // Start a new session
				}

				// Start child process
				if err := cmd.Start(); err != nil {
					return fmt.Errorf("failed to start background process: %w", err)
				}

				// Give the child process a moment to initialize
				time.Sleep(500 * time.Millisecond)

				// Check if the process is still running
				if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
					return fmt.Errorf("child process exited immediately - check logs at %s", logFilePath)
				}

				// Write PID to file
				pid := strconv.Itoa(cmd.Process.Pid)
				if err := os.WriteFile(pidFile, []byte(pid), 0644); err != nil {
					return fmt.Errorf("failed to write PID file: %w", err)
				}

				logger.Printf("Agent started in background with PID: %s (logs at %s)", pid, logFilePath)
				return nil
			}

			// Child process - continue execution
			logger.Println("Starting agent in background mode...")

			// Create a fresh context that is never cancelled
			ctx := context.Background()

			// Set up signal handling for clean shutdown
			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

			go func() {
				sig := <-sigCh
				logger.Printf("Received signal: %v, shutting down...", sig)

				// Clean up PID file
				os.Remove(pidFile)

				// Exit the process
				os.Exit(0)
			}()

			// Start the agent
			a := agent.New(wgIf, etcdCli, selfPubKey)

			// Run agent in a goroutine and monitor for errors
			errCh := make(chan error, 1)
			go func() {
				logger.Println("Starting agent.Run in background...")

				// Catch panics
				defer func() {
					if r := recover(); r != nil {
						logger.Printf("PANIC in Agent.Run: %v", r)
						errCh <- fmt.Errorf("agent panicked: %v", r)
					}
				}()

				// Run()がエラーを返す場合は、それをキャプチャする
				a.Run(ctx)

				// Agent.Run does not return an error, so just log and send nil
				logger.Println("Agent.Run returned (should not happen under normal operation)")

				// エラーかnilかにかかわらず、結果をチャネルに送信
				errCh <- nil
			}()

			// Block forever, but also monitor for agent errors
			logger.Println("Agent running in background mode")
			err := <-errCh
			if err != nil {
				logger.Printf("Agent stopped with error: %v", err)
				os.Remove(pidFile)
				return fmt.Errorf("agent stopped with error: %w", err)
			} else {
				logger.Println("Agent stopped unexpectedly without error")
				os.Remove(pidFile)
				return fmt.Errorf("agent stopped unexpectedly")
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(upCmd)
	upCmd.Flags().StringVar(&configPath, "config", "config.yaml", "Path to config file")
	upCmd.Flags().BoolVar(&debugMode, "debug", false, "Enable debug logging")
}
