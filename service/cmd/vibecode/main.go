package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ibrahim/remote-vibecode/cmd/vibecode/commands"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "vibecode",
		Short: "Vibecode - Remote tmux session management",
		Long: `Vibecode is a CLI tool for managing tmux sessions used with
the remote vibecode service. It provides an easy way to create, join,
list, and stop tmux sessions with custom configuration.`,
		Version: "1.0.0",
	}

	// Add subcommands
	rootCmd.AddCommand(commands.StartCmd)
	rootCmd.AddCommand(commands.JoinCmd)
	rootCmd.AddCommand(commands.ListCmd)
	rootCmd.AddCommand(commands.StopCmd)

	// Run the command
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
