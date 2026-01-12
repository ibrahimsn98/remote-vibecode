package session

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// TmuxSession represents a tmux session that the service is monitoring
type TmuxSession struct {
	ID          string
	SessionName string // tmux session name
	WindowName  string // window name (e.g., "main")
	PaneID      string // pane identifier
	Status      string // "attached", "detached", "dead"
	CreatedAt   time.Time
	LastCapture time.Time // last output capture time
	CapturePos  int       // position in tmux buffer for incremental capture (line count)

	mu       sync.RWMutex
	stopChan chan struct{} // channel to stop capture loop
}

// NewTmuxSession creates a new tmux session
func NewTmuxSession(sessionName, windowName string) *TmuxSession {
	return &TmuxSession{
		ID:          uuid.New().String(),
		SessionName: sessionName,
		WindowName:  windowName,
		Status:      "attached",
		CreatedAt:   time.Now(),
		CapturePos:  0, // Start from beginning
		stopChan:    make(chan struct{}),
	}
}

// GetID returns the session ID
func (s *TmuxSession) GetID() string {
	return s.ID
}

// GetStatus returns the current status
func (s *TmuxSession) GetStatus() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Status
}

// SetStatus updates the session status
func (s *TmuxSession) SetStatus(status string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Status = status
}

// GetCapturePosition returns the current capture position
func (s *TmuxSession) GetCapturePosition() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.CapturePos
}

// SetCapturePosition updates the capture position
func (s *TmuxSession) SetCapturePosition(pos int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CapturePos = pos
}

// IncrementCapturePosition increases the capture position
func (s *TmuxSession) IncrementCapturePosition(delta int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CapturePos += delta
}

// Stop stops the capture loop
func (s *TmuxSession) Stop() {
	close(s.stopChan)
}

// StopChan returns the stop channel
func (s *TmuxSession) StopChan() <-chan struct{} {
	return s.stopChan
}
