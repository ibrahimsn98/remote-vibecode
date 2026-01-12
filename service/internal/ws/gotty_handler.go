// Package ws provides gotty-based WebSocket handlers for terminal sharing
package ws

import (
	"encoding/base64"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/ibrahim/remote-vibecode/internal/gotty"
)

var gottyUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

const (
	// gotty protocol message types
	gottyOutput = '1'
	gottyInput  = '1'
	gottyPing   = '2'
	gottyPong   = '3'
	gottyResize = '4'
)

// GottyHandler handles gotty WebSocket connections for terminal sharing
type GottyHandler struct {
	gottyMgr *gotty.Manager
}

// NewGottyHandler creates a new gotty WebSocket handler
func NewGottyHandler(gottyMgr *gotty.Manager) *GottyHandler {
	return &GottyHandler{
		gottyMgr: gottyMgr,
	}
}

// HandleTmuxSession handles WebSocket connection for a tmux session using gotty protocol
// GET /gotty/:tmux_session
func (h *GottyHandler) HandleTmuxSession(c *gin.Context) {
	tmuxSessionName := c.Param("tmux_session")
	if tmuxSessionName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tmux_session parameter required"})
		return
	}

	// Attach to tmux session
	session, err := h.gottyMgr.AttachToTmuxSession(tmuxSessionName)
	if err != nil {
		log.Printf("Failed to attach to tmux session %s: %v", tmuxSessionName, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to attach to tmux session"})
		return
	}
	sessionID := session.ID

	// Upgrade to WebSocket
	conn, err := gottyUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade to WebSocket: %v", err)
		_ = session.Close()
		h.gottyMgr.RemoveSession(sessionID)
		return
	}

	log.Printf("Gotty session created: %s -> tmux:%s", sessionID[:8], tmuxSessionName)

	// Start bidirectional streaming
	var wg sync.WaitGroup
	wg.Add(2)

	// PTY -> WebSocket (output)
	go func() {
		defer wg.Done()
		buf := make([]byte, 4096)
		for {
			n, err := session.Read(buf)
			if err != nil {
				if !session.IsClosed() {
					log.Printf("PTY read error: %v", err)
				}
				break
			}

			// Encode as base64 and prefix with gottyOutput
			encoded := base64.StdEncoding.EncodeToString(buf[:n])
			msg := make([]byte, len(encoded)+1)
			msg[0] = gottyOutput
			copy(msg[1:], encoded)

			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				log.Printf("WebSocket write error: %v", err)
				break
			}
		}
	}()

	// WebSocket -> PTY (input)
	go func() {
		defer wg.Done()
		_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		conn.SetPongHandler(func(string) error {
			_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			return nil
		})

		for {
			messageType, data, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket read error: %v", err)
				}
				break
			}

			if messageType == websocket.TextMessage && len(data) > 0 {
				switch data[0] {
				case gottyInput:
					// User input - write directly to PTY
					if _, err := session.Write(data[1:]); err != nil {
						log.Printf("PTY write error: %v", err)
						break
					}

				case gottyPing:
					// Respond with pong
					_ = conn.WriteMessage(websocket.TextMessage, []byte{gottyPong})

				case gottyResize:
					// Resize request - format: columns,rows (as ASCII)
					// For now, ignore
					log.Printf("Resize request: %v", string(data[1:]))
				}
			}
		}
	}()

	wg.Wait()

	// Cleanup
	_ = session.Close()
	h.gottyMgr.RemoveSession(sessionID)
	_ = conn.Close()

	log.Printf("Gotty session closed: %s", sessionID[:8])
}
