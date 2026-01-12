package tmux

import (
	"fmt"
	"os/exec"
	"strings"
)

// SessionExists checks if a tmux session with the given name exists
func SessionExists(sessionName string) bool {
	cmd := exec.Command("tmux", "has-session", "-t", sessionName)
	return cmd.Run() == nil
}

// GetSessionInfo retrieves information about a tmux session
func GetSessionInfo(sessionName string) (map[string]string, error) {
	// Get session ID, window ID, and pane ID
	sessionID, err := exec.Command("tmux", "display-message", "-p", "-t", sessionName, "#{session_id}").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get session info: %w", err)
	}
	windowID, err := exec.Command("tmux", "display-message", "-p", "-t", sessionName, "#{window_id}").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get window info: %w", err)
	}
	paneID, err := exec.Command("tmux", "display-message", "-p", "-t", sessionName, "#{pane_id}").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get pane info: %w", err)
	}
	return map[string]string{
		"session_id": strings.TrimSpace(string(sessionID)),
		"window_id":  strings.TrimSpace(string(windowID)),
		"pane_id":    strings.TrimSpace(string(paneID)),
	}, nil
}

// ListSessions returns a list of all tmux session names
func ListSessions() ([]string, error) {
	cmd := exec.Command("tmux", "list-sessions", "-F", "#{session_name}")
	output, err := cmd.Output()
	if err != nil {
		// If no sessions exist, tmux returns error
		if strings.Contains(err.Error(), "no server running") || strings.Contains(err.Error(), "failed to") {
			return []string{}, nil
		}
		return nil, fmt.Errorf("list-sessions failed: %w", err)
	}

	sessions := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(sessions) == 1 && sessions[0] == "" {
		return []string{}, nil
	}
	return sessions, nil
}

// IsValidSessionName checks if a session name is valid (for security)
func IsValidSessionName(name string) bool {
	if name == "" {
		return false
	}

	// tmux session names should be alphanumeric, hyphen, underscore
	// Disallow special characters that could be used for command injection
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') ||
			c == '-' || c == '_') {
			return false
		}
	}
	return true
}
