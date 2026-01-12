package commands

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"

	"github.com/ibrahim/remote-vibecode/cmd/vibecode/internal/banner"
	"github.com/ibrahim/remote-vibecode/internal/tmux"
)

var StartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a new vibecode session",
	Long: `Start a new vibecode tmux session with custom configuration.
The session will have a distinctive status bar and startup banner.`,
	RunE: runStart,
}

func runStart(cmd *cobra.Command, args []string) error {
	// Check if tmux is installed
	if err := checkTmuxInstalled(); err != nil {
		return err
	}

	// Optional: Check if service is running (warn if not)
	checkServiceRunning()

	// Prompt for session name
	sessionName, err := promptSessionName()
	if err != nil {
		return err
	}

	// Create the session
	if err := tmux.CreateSession(sessionName); err != nil {
		return err
	}

	fmt.Printf("✓ Created vibecode session: %s\n", sessionName)

	// Set status bar
	if err := setStatusLine(sessionName); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to set status line: %v\n", err)
	}

	// Set shell prompt
	if err := setShellPrompt(sessionName); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to set shell prompt: %v\n", err)
	}

	// Display banner
	if err := displayBanner(sessionName); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to display banner: %v\n", err)
	}

	// Attach to session
	return attachSession(sessionName)
}

func checkTmuxInstalled() error {
	cmd := exec.Command("tmux", "-V")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("tmux not installed.\n\nInstall with:\n  macOS: brew install tmux\n  Linux (apt): sudo apt install tmux\n  Linux (yum): sudo yum install tmux\n  Linux (dnf): sudo dnf install tmux")
	}
	return nil
}

func checkServiceRunning() {
	// Try to connect to the service
	cmd := exec.Command("curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", "http://localhost:8080/api/v1/health")
	output, err := cmd.Output()
	if err != nil || string(output) != "200" {
		fmt.Fprintf(os.Stderr, "⚠ Warning: vibecode service may not be running.\n")
		fmt.Fprintf(os.Stderr, "  Start it with: brew services start remote-vibecode\n")
		fmt.Fprintf(os.Stderr, "  Or manually: cd service && go run cmd/server/main.go\n\n")
	}
}

func promptSessionName() (string, error) {
	defaultName := fmt.Sprintf("vibecode-%d", time.Now().Unix())

	var sessionName string
	prompt := &survey.Input{
		Message: "Session name:",
		Default: defaultName,
	}
	err := survey.AskOne(prompt, &sessionName, survey.WithValidator(survey.Required))
	if err != nil {
		return "", fmt.Errorf("prompt failed: %w", err)
	}

	// Validate session name
	if !tmux.IsValidSessionName(sessionName) {
		return "", fmt.Errorf("invalid session name: must contain only letters, numbers, hyphens, and underscores")
	}

	// Check if session already exists
	if tmux.SessionExists(sessionName) {
		return "", fmt.Errorf("session '%s' already exists. Use 'vibecode join %s' to connect.", sessionName, sessionName)
	}

	return sessionName, nil
}

func setStatusLine(sessionName string) error {
	// Set the global status line options (not session-specific)
	statusConfig := []string{
		"set-option -g status-left '#[bg=green]#[fg=black] VIBE #[default] #{session_name} '",
		"set-option -g status-right '%H:%M %d-%b-%y'",
		"set-option -g status-bg '#1a1a2e'",
		"set-option -g status-fg '#eee8aa'",
		"set-option -g status-interval 1",
	}

	for _, cfg := range statusConfig {
		cmd := exec.Command("tmux", "set-option", "-g")
		// Parse the config string into arguments
		// set-option -g status-left '...'
		// becomes: ["set-option", "-g", "status-left", "..."]
		args := []string{"set-option", "-g"}
		parts := quoteSplit(cfg[len("set-option -g "):])
		args = append(args, parts...)

		cmd = exec.Command("tmux", args...)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to run: tmux %s: %w", cfg, err)
		}
	}
	return nil
}

func setShellPrompt(sessionName string) error {
	// Detect shell and set appropriate prompt
	// For bash/zsh: set PS1 with VIBE prefix
	// Using ANSI colors: green for VIBE, normal for rest of prompt

	// Try to detect the shell by checking SHELL env var
	detectCmd := `echo $SHELL`
	shell := "zsh" // default to zsh on macOS

	cmd := exec.Command("sh", "-c", detectCmd)
	if output, err := cmd.Output(); err == nil {
		shellPath := string(output)
		if strings.Contains(shellPath, "bash") {
			shell = "bash"
		} else if strings.Contains(shellPath, "zsh") {
			shell = "zsh"
		} else if strings.Contains(shellPath, "fish") {
			shell = "fish"
		}
	}

	var promptCmd string
	switch shell {
	case "bash":
		// Bash PS1 with VIBE prefix
		promptCmd = `export PS1="\[\033[1;32m\]VIBE\[\033[0m\] \w $ "`
	case "zsh":
		// Zsh prompt with VIBE prefix
		promptCmd = `export PS1="%F{green}VIBE%f %1~ %# "`
	case "fish":
		// Fish prompt
		promptCmd = `function fish_prompt; echo -n (set_color green)"VIBE"(set_color normal)" "(prompt_pwd)"> "; end`
	default:
		// Generic
		promptCmd = `export PS1="VIBE \\w $ "`
	}

	// Add the prompt command to shell startup
	// We'll write it to a temp file and source it
	rcFile := "/tmp/vibecode-prompt-" + sessionName + ".sh"
	rcContent := "#!/bin/sh\n" + promptCmd + "\n"

	if err := os.WriteFile(rcFile, []byte(rcContent), 0644); err != nil {
		return err
	}

	// Source the rc file in the session
	sourceCmd := "source " + rcFile + " && rm " + rcFile
	if err := tmux.SendCommand(sessionName, sourceCmd); err != nil {
		return err
	}

	return nil
}

func displayBanner(sessionName string) error {
	// Create a temporary file with the banner content
	tmpFile := "/tmp/vibecode-banner-" + sessionName + ".txt"
	bannerText := banner.String(sessionName)

	// Write banner to temp file
	if err := os.WriteFile(tmpFile, []byte(bannerText), 0644); err != nil {
		return fmt.Errorf("failed to write banner file: %w", err)
	}

	// Send commands to clear screen and cat the banner
	if err := tmux.SendCommand(sessionName, "clear"); err != nil {
		return err
	}
	if err := tmux.SendCommand(sessionName, "cat "+tmpFile+" && rm "+tmpFile); err != nil {
		return err
	}

	return nil
}

// quoteSplit splits a string by spaces, respecting quotes
func quoteSplit(s string) []string {
	var result []string
	var current string
	var inQuote bool
	var quoteChar rune

	for _, r := range s {
		switch {
		case (r == '\'' || r == '"') && !inQuote:
			inQuote = true
			quoteChar = r
		case r == quoteChar && inQuote:
			inQuote = false
		case r == ' ' && !inQuote:
			if current != "" {
				result = append(result, current)
				current = ""
			}
		default:
			current += string(r)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}
