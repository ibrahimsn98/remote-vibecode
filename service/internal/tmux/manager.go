package tmux

import (
	"log"
	"strings"
	"sync"
	"time"

	"github.com/ibrahim/remote-vibecode/internal/session"
	"github.com/ibrahim/remote-vibecode/internal/ws"
)

// Manager manages tmux session tracking and discovery
type Manager struct {
	mu                 sync.RWMutex
	sessions           map[string]*session.TmuxSession // session ID -> TmuxSession
	sessionByName      map[string]*session.TmuxSession // session name -> TmuxSession
	discoveryInterval  time.Duration
	stopDiscovery      chan struct{}
	autoAttachPatterns []string       // Session name patterns to auto-attach (e.g., "claude", "tmux-*")
	sessionHub         *ws.SessionHub // Hub for broadcasting session updates
}

// New creates a new tmux manager
func New(sessionHub *ws.SessionHub) *Manager {
	m := &Manager{
		sessions:           make(map[string]*session.TmuxSession),
		sessionByName:      make(map[string]*session.TmuxSession),
		discoveryInterval:  2 * time.Second, // Check for new sessions every 2 seconds
		stopDiscovery:      make(chan struct{}),
		autoAttachPatterns: []string{"claude", "tmux"}, // Auto-attach to sessions starting with these
		sessionHub:         sessionHub,
	}

	// Start auto-discovery loop
	go m.discoveryLoop()

	return m
}

// AttachSession attaches to an existing tmux session (for tracking only)
func (m *Manager) AttachSession(sessionName string) (*session.TmuxSession, error) {
	// Validate session name
	if !IsValidSessionName(sessionName) {
		return nil, ErrInvalidSessionName
	}

	// Check if session exists
	if !SessionExists(sessionName) {
		return nil, ErrSessionNotFound
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if already attached
	if sess, exists := m.sessionByName[sessionName]; exists {
		return sess, nil
	}

	// Get session info
	info, err := GetSessionInfo(sessionName)
	if err != nil {
		return nil, err
	}

	// Create new session
	sess := session.NewTmuxSession(sessionName, info["window_id"])
	sess.PaneID = info["pane_id"]

	// Store session
	m.sessions[sess.ID] = sess
	m.sessionByName[sessionName] = sess

	log.Printf("Tracking tmux session: %s (id: %s)", sessionName, sess.ID)

	return sess, nil
}

// DetachSession detaches from a tmux session
func (m *Manager) DetachSession(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	sess, exists := m.sessions[sessionID]
	if !exists {
		return
	}

	sess.Stop()
	sess.SetStatus("detached")

	// Remove from maps
	delete(m.sessions, sessionID)
	delete(m.sessionByName, sess.SessionName)

	log.Printf("Stopped tracking tmux session: %s", sess.SessionName)
}

// GetSession retrieves a session by ID
func (m *Manager) GetSession(sessionID string) (*session.TmuxSession, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sess, exists := m.sessions[sessionID]
	return sess, exists
}

// GetSessionByName retrieves a session by tmux session name
func (m *Manager) GetSessionByName(sessionName string) (*session.TmuxSession, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sess, exists := m.sessionByName[sessionName]
	return sess, exists
}

// ListSessions returns all tracked tmux sessions
func (m *Manager) ListSessions() []*session.TmuxSession {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*session.TmuxSession, 0, len(m.sessions))
	for _, sess := range m.sessions {
		result = append(result, sess)
	}

	return result
}

// Errors
var (
	ErrInvalidSessionName = &TmuxError{Message: "invalid session name"}
	ErrSessionNotFound    = &TmuxError{Message: "tmux session not found"}
)

// TmuxError represents a tmux-related error
type TmuxError struct {
	Message string
}

func (e *TmuxError) Error() string {
	return e.Message
}

// discoveryLoop periodically scans for new tmux sessions and auto-attaches
func (m *Manager) discoveryLoop() {
	ticker := time.NewTicker(m.discoveryInterval)
	defer ticker.Stop()

	// Do initial scan on startup
	m.scanAndAttach()

	for {
		select {
		case <-m.stopDiscovery:
			return
		case <-ticker.C:
			m.scanAndAttach()
		}
	}
}

// scanAndAttach scans for tmux sessions and attaches to matching ones
func (m *Manager) scanAndAttach() {
	// Get all tmux sessions
	sessionNames, err := ListSessions()
	if err != nil {
		return
	}

	// Check each session
	for _, sessionName := range sessionNames {
		// Skip if already attached
		m.mu.RLock()
		_, alreadyAttached := m.sessionByName[sessionName]
		m.mu.RUnlock()

		if alreadyAttached {
			continue
		}

		// Check if session matches auto-attach patterns
		if m.shouldAutoAttach(sessionName) {
			log.Printf("Auto-discovered tmux session: %s", sessionName)
			m.AttachSession(sessionName)
		}
	}

	// Broadcast updated session list to all connected clients
	if m.sessionHub != nil {
		m.broadcastSessions()
	}
}

// broadcastSessions sends the current session list to all connected WebSocket clients
func (m *Manager) broadcastSessions() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sessionInfos := make([]ws.SessionInfo, 0, len(m.sessions))
	for _, sess := range m.sessions {
		sessionInfos = append(sessionInfos, ws.SessionInfo{
			ID:          sess.ID,
			SessionName: sess.SessionName,
			CreatedAt:   sess.CreatedAt.Unix(),
			LastCapture: sess.LastCapture.Unix(),
		})
	}

	m.sessionHub.BroadcastSessions(sessionInfos)
}

// shouldAutoAttach checks if a session name matches auto-attach patterns
func (m *Manager) shouldAutoAttach(sessionName string) bool {
	for _, pattern := range m.autoAttachPatterns {
		// Check if session name starts with pattern
		if strings.HasPrefix(sessionName, pattern) {
			return true
		}
	}
	return false
}

// Stop stops the discovery loop
func (m *Manager) Stop() {
	close(m.stopDiscovery)
}
