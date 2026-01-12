package commands

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"

	"github.com/ibrahim/remote-vibecode/cmd/vibecode/internal/banner"
	"github.com/ibrahim/remote-vibecode/internal/tmux"
)

var StartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a new rvc session",
	Long: `Start a new rvc tmux session with custom configuration.
The session will have a distinctive status bar and startup banner.`,
	RunE: runStart,
}

var writableFlag bool

func init() {
	StartCmd.Flags().BoolVarP(&writableFlag, "writable", "w", false, "Create a writable session (web clients can type)")
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

	// Set writable flag if -w was provided
	if writableFlag {
		if err := tmux.SetWritable(sessionName, true); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to set writable flag: %v\n", err)
		}
		fmt.Printf("Session '%s' is WRITABLE (web clients can type)\n", sessionName)
	} else {
		fmt.Printf("Session '%s' is READ-ONLY (use -w flag for writable)\n", sessionName)
	}

	// Set status bar
	if err := setStatusLine(sessionName); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to set status line: %v\n", err)
	}

	// Show banner
	fmt.Println()
	fmt.Print(banner.String(sessionName))

	// Show progress and wait
	fmt.Printf("\n► Starting %s rvc session", sessionName)
	for i := 0; i < 5; i++ {
		time.Sleep(1000 * time.Millisecond)
		fmt.Print(".")
	}
	fmt.Println(" ✓")

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
		fmt.Fprintf(os.Stderr, "⚠ Warning: rvc servcer may not be running.\n")
		fmt.Fprintf(os.Stderr, "  Start it with: rvc servce\n\n")
	}
}

func promptSessionName() (string, error) {
	defaultName := fmt.Sprintf("rvc-%d", time.Now().Unix())

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
		return "", fmt.Errorf("session '%s' already exists. Use 'rvc join %s' to connect.", sessionName, sessionName)
	}

	return sessionName, nil
}

func setStatusLine(sessionName string) error {
	// Set the global status line options (not session-specific)
	statusConfig := []string{
		"set-option -g status-left '#[bg=green]#[fg=black] RV #[default] #{session_name} '",
		"set-option -g status-right '%H:%M %d-%b-%y'",
		"set-option -g status-bg '#1a1a2e'",
		"set-option -g status-fg '#eee8aa'",
		"set-option -g status-intervcal 1",
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
