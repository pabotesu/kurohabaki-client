package util

import (
	"os"
	"path/filepath"
	"runtime"
)

// GetPidFilePath returns the appropriate path for the PID file
// based on the current OS and user permissions
func GetPidFilePath() string {
	// For Windows
	if runtime.GOOS == "windows" {
		// Use %TEMP% directory on Windows
		return filepath.Join(os.TempDir(), "kh-client.pid")
	}

	// For Unix-like systems (Linux, macOS)
	// Check if we're running as root
	if os.Geteuid() == 0 {
		// Standard location for system daemons
		return "/var/run/kh-client.pid"
	}

	// For non-root users on Unix systems
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to temporary directory if home directory is unavailable
		return filepath.Join(os.TempDir(), "kh-client.pid")
	}

	// Use hidden file in user's home directory
	return filepath.Join(homeDir, ".kh-client.pid")
}
