package commands

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"

	"github.com/ibrahim/remote-vibecode/internal/tmux"
)

var (
	forceStop bool
)

var StopCmd = &cobra.Command{
	Use:   "stop [session-name]",
	Short: "Stop an rv session",
	Long:  `Stop (kill) an rv tmux session. You'll be prompted to confirm unless --force is used.`,
	RunE:  runStop,
}

func init() {
	StopCmd.Flags().BoolVarP(&forceStop, "force", "f", false, "Skip confirmation prompt")
}

func runStop(cmd *cobra.Command, args []string) error {
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
		return fmt.Errorf("session '%s' does not exist. Use 'rv list' to see available sessions.", sessionName)
	}

	// Confirm before killing (unless --force)
	if !forceStop {
		confirmed := false
		prompt := &survey.Confirm{
			Message: fmt.Sprintf("Are you sure you want to stop session '%s'?", sessionName),
			Default: false,
		}
		if err := survey.AskOne(prompt, &confirmed); err != nil {
			return err
		}
		if !confirmed {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	// Kill the session
	if err := tmux.KillSession(sessionName); err != nil {
		return err
	}

	fmt.Printf("✓ Stopped session: %s\n", sessionName)
	return nil
}
