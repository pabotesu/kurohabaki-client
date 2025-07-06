package cmd

import (
	"fmt"
	"os"
	"strconv"
	"syscall"

	"github.com/pabotesu/kurohabaki-client/internal/logger"
	"github.com/spf13/cobra"
)

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Stop the running WireGuard interface and agent",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if PID file exists
		pidFile := "/var/run/kh-client.pid"
		pidData, err := os.ReadFile(pidFile)
		if err != nil {
			return fmt.Errorf("no running agent found: %w", err)
		}

		// Parse PID
		pid, err := strconv.Atoi(string(pidData))
		if err != nil {
			return fmt.Errorf("invalid PID in pidfile: %w", err)
		}

		// Find the process
		process, err := os.FindProcess(pid)
		if err != nil {
			return fmt.Errorf("failed to find process with PID %d: %w", pid, err)
		}

		// Send SIGTERM to gracefully shut down the agent
		logger.Println("Sending shutdown signal to agent (PID: " + strconv.Itoa(pid) + ")...")
		if err := process.Signal(syscall.SIGTERM); err != nil {
			return fmt.Errorf("failed to send shutdown signal: %w", err)
		}

		// Clean up PID file
		if err := os.Remove(pidFile); err != nil {
			logger.Printf("Warning: failed to remove PID file: %v", err)
		}

		logger.Println("Agent stopped successfully")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(downCmd)
}
