package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/pabotesu/kurohabaki-client/internal/logger"
	"github.com/pabotesu/kurohabaki-client/internal/util"
	"github.com/spf13/cobra"
)

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Stop the running WireGuard interface and agent",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get the appropriate PID file path based on OS and permissions
		pidFile := util.GetPidFilePath()

		// Check if PID file exists
		if _, err := os.Stat(pidFile); os.IsNotExist(err) {
			return fmt.Errorf("no running agent found: PID file does not exist")
		}

		// Read PID file
		pidData, err := os.ReadFile(pidFile)
		if err != nil {
			return fmt.Errorf("failed to read PID file: %w", err)
		}

		// Parse PID - trim any whitespace
		pidStr := strings.TrimSpace(string(pidData))
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			// Remove invalid PID file
			os.Remove(pidFile)
			return fmt.Errorf("invalid PID in pidfile (file removed): %w", err)
		}

		// Find the process
		process, err := os.FindProcess(pid)
		if err != nil {
			// Remove PID file if process not found
			os.Remove(pidFile)
			return fmt.Errorf("failed to find process with PID %d (file removed): %w", pid, err)
		}

		// Send SIGTERM to gracefully shut down the agent
		logger.Printf("Sending shutdown signal to agent (PID: %d)...", pid)
		if err := process.Signal(syscall.SIGTERM); err != nil {
			// If signal sending fails, process is likely already gone
			os.Remove(pidFile)
			logger.Println("Process not running, removed PID file")
			return nil
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
