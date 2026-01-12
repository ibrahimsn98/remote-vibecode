// Package gotty provides terminal sharing using gotty protocol
package gotty

import (
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/creack/pty"
	"github.com/google/uuid"
)

// Session represents an active gotty terminal session
type Session struct {
	ID       string
	TmuxName string
	Cmd      *exec.Cmd
	Pty      *os.File
	mu       sync.RWMutex
	closed   bool
}

// Manager manages gotty terminal sessions
type Manager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

// NewManager creates a new gotty session manager
func NewManager() *Manager {
	return &Manager{
		sessions: make(map[string]*Session),
	}
}

// AttachToTmuxSession attaches to an existing tmux session
func (m *Manager) AttachToTmuxSession(tmuxSessionName string) (*Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create command to attach to tmux session
	cmd := exec.Command("tmux", "attach", "-t", tmuxSessionName)

	// Start PTY
	ptyFile, err := pty.Start(cmd)
	if err != nil {
		return nil, err
	}

	session := &Session{
		ID:       uuid.New().String(),
		TmuxName: tmuxSessionName,
		Cmd:      cmd,
		Pty:      ptyFile,
		closed:   false,
	}

	m.sessions[session.ID] = session
	return session, nil
}

// GetSession retrieves a session by ID
func (m *Manager) GetSession(id string) (*Session, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sess, ok := m.sessions[id]
	return sess, ok
}

// RemoveSession removes a session from the manager
func (m *Manager) RemoveSession(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.sessions, id)
}

// ListSessions returns all active sessions
func (m *Manager) ListSessions() []*Session {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sessions := make([]*Session, 0, len(m.sessions))
	for _, sess := range m.sessions {
		sessions = append(sessions, sess)
	}
	return sessions
}

// Close closes the session
func (s *Session) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}
	s.closed = true

	// Close PTY
	if err := s.Pty.Close(); err != nil {
		return err
	}

	// Kill process
	if s.Cmd.Process != nil {
		return s.Cmd.Process.Kill()
	}

	return nil
}

// IsClosed returns whether the session is closed
func (s *Session) IsClosed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.closed
}

// Read reads from the PTY
func (s *Session) Read(p []byte) (n int, err error) {
	if s.IsClosed() {
		return 0, io.EOF
	}
	return s.Pty.Read(p)
}

// Write writes to the PTY
func (s *Session) Write(p []byte) (n int, err error) {
	if s.IsClosed() {
		return 0, io.ErrClosedPipe
	}
	return s.Pty.Write(p)
}
