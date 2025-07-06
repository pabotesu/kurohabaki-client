package cmd

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pabotesu/kurohabaki-client/config"
	"github.com/pabotesu/kurohabaki-client/internal/agent"
	"github.com/pabotesu/kurohabaki-client/internal/etcd"
	"github.com/pabotesu/kurohabaki-client/internal/logger"
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
			Logger:      zapLogger, // Add this line
		})
		if err != nil {
			return fmt.Errorf("failed to connect to etcd: %w", err)
		}
		defer etcdCli.Close()

		// Check etcd health
		if err := etcd.CheckEtcdHealth(etcdCli); err != nil {
			logger.Println("⚠️ Warning: " + err.Error())
			logger.Println("⚠️ Will continue but peer synchronization may not work")
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

		// Graceful shutdown on SIGINT/SIGTERM
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		a := agent.New(wgIf, etcdCli, selfPubKey)

		// Handle signals for graceful shutdown
		go func() {
			sig := <-sigCh
			logger.Printf("🛑 Caught signal: %v, shutting down...", sig)
			cancel()
		}()

		// Start the agent
		a.Run(ctx)

		logger.Println("🏁 Agent stopped, exiting normally.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(upCmd)
	upCmd.Flags().StringVar(&configPath, "config", "config.yaml", "Path to config file")
	upCmd.Flags().BoolVar(&debugMode, "debug", false, "Enable debug logging")
}
