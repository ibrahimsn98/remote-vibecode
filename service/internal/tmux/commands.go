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

// CreateSession creates a new tmux session with the given name
func CreateSession(sessionName string) error {
	return CreateSessionWithPrompt(sessionName, "")
}

// CreateSessionWithPrompt creates a new tmux session with a custom prompt
// The prompt is set via shell initialization, avoiding visible commands
func CreateSessionWithPrompt(sessionName, prompt string) error {
	if !IsValidSessionName(sessionName) {
		return fmt.Errorf("invalid session name: %s", sessionName)
	}
	if SessionExists(sessionName) {
		return fmt.Errorf("session '%s' already exists", sessionName)
	}

	// Detect shell
	shell := "zsh"
	detectCmd := exec.Command("sh", "-c", "echo $SHELL")
	if output, err := detectCmd.Output(); err == nil {
		shellPath := strings.TrimSpace(string(output))
		if strings.Contains(shellPath, "bash") {
			shell = "bash"
		} else if strings.Contains(shellPath, "zsh") {
			shell = "zsh"
		} else if strings.Contains(shellPath, "fish") {
			shell = "fish"
		}
	}

	// Build initial command that sets prompt and starts shell
	var initialCmd string
	if prompt != "" {
		switch shell {
		case "bash":
			initialCmd = fmt.Sprintf("bash -c 'PS1=%q; exec bash'", prompt)
		case "zsh":
			initialCmd = fmt.Sprintf("zsh -c 'PS1=%q; exec zsh'", prompt)
		case "fish":
			// Fish uses a different prompt system - use environment variable approach
			fishPrompt := `(set_color green)"VIBE"(set_color normal) " "(prompt_pwd)"> "`
			initialCmd = fmt.Sprintf("fish -c 'function fish_prompt; echo -n %s; end; exec fish'", fishPrompt)
		default:
			initialCmd = shell
		}
	} else {
		initialCmd = shell
	}

	// Create session with initial command
	cmd := exec.Command("tmux", "new-session", "-d", "-s", sessionName, initialCmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create session: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// KillSession kills (deletes) a tmux session
func KillSession(sessionName string) error {
	if !IsValidSessionName(sessionName) {
		return fmt.Errorf("invalid session name: %s", sessionName)
	}
	if !SessionExists(sessionName) {
		return fmt.Errorf("session '%s' does not exist", sessionName)
	}
	cmd := exec.Command("tmux", "kill-session", "-t", sessionName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to kill session: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// AttachSession attaches the current terminal to an existing tmux session
// This should be called when the process will replace itself with tmux
func AttachSession(sessionName string) error {
	if !IsValidSessionName(sessionName) {
		return fmt.Errorf("invalid session name: %s", sessionName)
	}
	if !SessionExists(sessionName) {
		return fmt.Errorf("session '%s' does not exist", sessionName)
	}
	// This returns an error because we intend to exec it, not run it
	return fmt.Errorf("use exec.Command for tmux attach")
}

// SendCommand sends a command to a tmux session (e.g., to set status bar or display banner)
func SendCommand(sessionName, command string) error {
	if !IsValidSessionName(sessionName) {
		return fmt.Errorf("invalid session name: %s", sessionName)
	}
	if !SessionExists(sessionName) {
		return fmt.Errorf("session '%s' does not exist", sessionName)
	}
	cmd := exec.Command("tmux", "send-keys", "-t", sessionName, command, "Enter")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to send command to session: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// SetStatusLine sets the tmux status line configuration for a session
func SetStatusLine(sessionName string) error {
	statusConfig := []string{
		"set-option -g status-left '#[bg=green]#[fg=black] RVC #[default] #{session_name} '",
		"set-option -g status-right '%H:%M %d-%b-%y'",
		"set-option -g status-bg '#1a1a2e'",
		"set-option -g status-fg '#eee8aa'",
		"set-option -g status-interval 1",
	}

	for _, cfg := range statusConfig {
		cmd := exec.Command("tmux", strings.Fields(cfg)...)
		cmd.Args = append([]string{"tmux"}, strings.Fields(cfg)...)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to set status line: %w", err)
		}
	}
	return nil
}

// SetWritable marks a session as writable or read-only using tmux user-options
func SetWritable(sessionName string, writable bool) error {
	value := "0"
	if writable {
		value = "1"
	}
	cmd := exec.Command("tmux", "set-option", "-t", sessionName, "@rvc-writable", value)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set writable flag: %w", err)
	}
	return nil
}

// IsWritable checks if a session is writable (returns false if not set)
func IsWritable(sessionName string) bool {
	cmd := exec.Command("tmux", "show-option", "-t", sessionName, "-qv", "@rvc-writable")
	output, err := cmd.Output()
	if err != nil {
		return false // Not set = read-only
	}
	return strings.TrimSpace(string(output)) == "1"
}
