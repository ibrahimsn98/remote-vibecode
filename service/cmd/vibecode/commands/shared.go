package commands

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/AlecAivazis/survey/v2"
	"github.com/ibrahim/remote-vibecode/internal/tmux"
)

// attachSession attaches the current terminal to a tmux session
func attachSession(sessionName string) error {
	cmd := exec.Command("tmux", "attach-session", "-t", sessionName)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("\nSession continues running after you detach.\n")
	fmt.Printf("Rejoin with: vibecode join %s\n\n", sessionName)

	return cmd.Run()
}

// promptSessionSelection prompts the user to select a session from the list
func promptSessionSelection() (string, error) {
	sessions, err := tmux.ListSessions()
	if err != nil {
		return "", fmt.Errorf("failed to list sessions: %w", err)
	}

	if len(sessions) == 0 {
		return "", fmt.Errorf("no tmux sessions found")
	}

	var sessionName string
	prompt := &survey.Select{
		Message: "Select a session:",
		Options: sessions,
	}
	err = survey.AskOne(prompt, &sessionName)
	if err != nil {
		return "", fmt.Errorf("prompt failed: %w", err)
	}

	return sessionName, nil
}
