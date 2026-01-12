package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ibrahim/remote-vibecode/internal/tmux"
)

var JoinCmd = &cobra.Command{
	Use:   "join [session-name]",
	Short: "Join an existing vibecode session",
	Long: `Join an existing vibecode tmux session.
If no session name is provided, you'll be prompted to select from available sessions.`,
	RunE: runJoin,
}

func runJoin(cmd *cobra.Command, args []string) error {
	var sessionName string

	if len(args) > 0 {
		sessionName = args[0]
	} else {
		// Prompt for session selection
		var err error
		sessionName, err = promptSessionSelection()
		if err != nil {
			return err
		}
	}

	// Validate session name
	if !tmux.IsValidSessionName(sessionName) {
		return fmt.Errorf("invalid session name: %s", sessionName)
	}

	// Check if session exists
	if !tmux.SessionExists(sessionName) {
		return fmt.Errorf("session '%s' does not exist. Use 'vibecode list' to see available sessions.", sessionName)
	}

	// Attach to session
	return attachSession(sessionName)
}
