package commands

import (
	"fmt"
	"os/exec"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ibrahim/remote-vibecode/internal/tmux"
)

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all rv sessions",
	Long:  `List all available tmux sessions with their details.`,
	RunE:  runList,
}

func runList(cmd *cobra.Command, args []string) error {
	sessions, err := tmux.ListSessions()
	if err != nil {
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	if len(sessions) == 0 {
		fmt.Println("No tmux sessions found.")
		fmt.Println("Create one with: rv start")
		return nil
	}

	// Use tabwriter for nice formatting
	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "SESSION NAME\tWINDOWS\tCREATED")

	for _, sessionName := range sessions {
		// Get session info
		info, err := getSessionDetails(sessionName)
		if err != nil {
			fmt.Fprintf(w, "%s\t?\t?\n", sessionName)
			continue
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", sessionName, info.windows, info.created)
	}

	w.Flush()
	fmt.Printf("\nTotal sessions: %d\n", len(sessions))
	fmt.Println("Join a session: rv join [session-name]")

	return nil
}

type sessionInfo struct {
	windows string
	created string
}

func getSessionDetails(sessionName string) (*sessionInfo, error) {
	// Get window count
	cmd := exec.Command("tmux", "display-message", "-p", "-t", sessionName, "#{session_windows}")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	windows := strings.TrimSpace(string(output))

	// Get creation time (using session group format)
	// Note: tmux doesn't easily expose creation time, so we'll show the session group instead
	cmd = exec.Command("tmux", "display-message", "-p", "-t", sessionName, "#{session_group}")
	output, err = cmd.Output()
	if err != nil {
		return nil, err
	}
	group := strings.TrimSpace(string(output))

	return &sessionInfo{
		windows: windows,
		created: group,
	}, nil
}
