package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kurohabaki",
	Short: "Kurohabaki client CLI",
	Long:  `Kurohabaki is a lightweight WireGuard-based P2P networking client.`,
}

func init() {
	// vewrsion flag
	rootCmd.Version = "0.1.0"
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
