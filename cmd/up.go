package cmd

import (
	"fmt"
	"log"

	"github.com/pabotesu/kurohabaki-client/config"
	"github.com/pabotesu/kurohabaki-client/internal/wg"
	"github.com/spf13/cobra"
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

		if err := wgIf.DumpConfig(); err != nil {
			log.Printf("Failed to dump config: %v", err)
		}

		select {}

	},
}

func init() {
	rootCmd.AddCommand(upCmd)
	upCmd.Flags().StringVar(&configPath, "config", "config.yaml", "Path to config file")
}
