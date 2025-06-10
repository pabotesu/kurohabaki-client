package cmd

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"github.com/pabotesu/kurohabaki-client/config"
	"github.com/pabotesu/kurohabaki-client/internal/agent"
	"github.com/pabotesu/kurohabaki-client/internal/wg"
	"github.com/spf13/cobra"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var configPath string

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Start WireGuard interface and connect to peers",
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Println("Bringing up WireGuard interface...")

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
		log.Println("WireGuard interface is up")
		// Prevent process from exiting to keep interface alive

		etcdCli, err := clientv3.New(clientv3.Config{
			Endpoints:   []string{cfg.Etcd.Endpoint},
			DialTimeout: 5 * time.Second,
		})
		if err != nil {
			return fmt.Errorf("failed to connect to etcd: %w", err)
		}
		defer etcdCli.Close()

		selfPubKey := base64.StdEncoding.EncodeToString(conf.Peers[0].PublicKey[:])
		log.Printf("🔎 Peer count in conf: %d", len(conf.Peers))
		log.Printf("✅ Peers in config: %d", len(conf.Peers))
		log.Printf("✅ PublicKey (self): %s", selfPubKey)
		log.Printf("✅ etcd endpoint: %s", cfg.Etcd.Endpoint)
		log.Println("✅ Starting Agent...")

		a := agent.New(wgIf, etcdCli, selfPubKey)
		a.Run(context.Background())
		log.Println("🟢 Agent.Run started")
		log.Println("🟢 Launching StartPeerWatcher goroutine")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(upCmd)
	upCmd.Flags().StringVar(&configPath, "config", "config.yaml", "Path to config file")
}
