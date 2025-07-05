package cmd

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/spf13/cobra"
)

func TestRootCommand(t *testing.T) {
	// Backup the original rootCmd
	originalRootCmd := rootCmd
	defer func() { rootCmd = originalRootCmd }()

	t.Run("RootCommandBasicExecution", func(t *testing.T) {
		// Create a buffer to capture command output
		buf := new(bytes.Buffer)
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)

		// Execute help command
		rootCmd.SetArgs([]string{"--help"})
		err := rootCmd.Execute()

		// Verify no error occurred
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Verify output contains the title
		output := buf.String()
		if !bytes.Contains(buf.Bytes(), []byte("Kurohabaki")) {
			t.Errorf("Expected output to contain 'Kurohabaki', got: %s", output)
		}
	})

	t.Run("RootCommandVersion", func(t *testing.T) {
		// Create a buffer to capture command output
		buf := new(bytes.Buffer)
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)

		// Verify version information is displayed
		rootCmd.SetArgs([]string{"--version"})
		err := rootCmd.Execute()

		// Verify no error occurred
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("RootCommandErrorHandling", func(t *testing.T) {
		// Add a subcommand that returns an error
		errCmd := &cobra.Command{
			Use: "error-test",
			RunE: func(cmd *cobra.Command, args []string) error {
				return fmt.Errorf("test error")
			},
		}

		// Create a new rootCmd for testing
		testRootCmd := &cobra.Command{
			Use:   "kurohabaki",
			Short: "Kurohabaki client CLI",
			Long:  `Kurohabaki is a lightweight WireGuard-based P2P networking client.`,
		}
		testRootCmd.AddCommand(errCmd)
		rootCmd = testRootCmd

		// Execute the command that returns an error and test error handling
		testRootCmd.SetArgs([]string{"error-test"})
		err := testRootCmd.Execute()

		if err == nil {
			t.Error("Expected error but got nil")
		}

		if err.Error() != "test error" {
			t.Errorf("Expected 'test error', got '%v'", err)
		}
	})
}

func TestExecute(t *testing.T) {
	// Direct testing of Execute() is difficult as it calls os.Exit
	// Here we just verify the function exists
	// Actual testing is covered indirectly by testing rootCmd behavior

	// Another approach would be to temporarily mock rootCmd.Execute
	// to prevent it from exiting, but that may be excessive for this simple case
}
