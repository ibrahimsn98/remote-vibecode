package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/ibrahim/remote-vibecode/internal/tmux"
	"github.com/ibrahim/remote-vibecode/internal/ws"
)

var sessionsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// TmuxHandlers provides tmux-related API endpoints
type TmuxHandlers struct {
	manager    *tmux.Manager
	sessionHub *ws.SessionHub
}

// NewTmuxHandlers creates a new tmux handlers instance
func NewTmuxHandlers(manager *tmux.Manager, sessionHub *ws.SessionHub) *TmuxHandlers {
	return &TmuxHandlers{
		manager:    manager,
		sessionHub: sessionHub,
	}
}

// ListSessions lists all active tmux sessions
// GET /api/v1/tmux/sessions
func (h *TmuxHandlers) ListSessions(c *gin.Context) {
	sessions := h.manager.ListSessions()

	result := make([]map[string]interface{}, 0, len(sessions))
	for _, sess := range sessions {
		result = append(result, map[string]interface{}{
			"id":           sess.ID,
			"session_name": sess.SessionName,
			"window_name":  sess.WindowName,
			"pane_id":      sess.PaneID,
			"status":       sess.GetStatus(),
			"created_at":   sess.CreatedAt,
			"last_capture": sess.LastCapture,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"sessions": result,
	})
}

// SessionWebSocket handles WebSocket connection for session list updates
// GET /api/v1/sessions/ws
func (h *TmuxHandlers) SessionWebSocket(c *gin.Context) {
	conn, err := sessionsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade to WebSocket: %v", err)
		return
	}

	client := &ws.SessionClient{
		Conn: conn,
		Hub:  h.sessionHub,
		Send: make(chan []byte, 256),
	}

	h.sessionHub.Register(client)

	// Send current sessions immediately
	sessions := h.manager.ListSessions()
	sessionInfos := make([]ws.SessionInfo, 0, len(sessions))
	for _, sess := range sessions {
		sessionInfos = append(sessionInfos, ws.SessionInfo{
			ID:          sess.ID,
			SessionName: sess.SessionName,
			CreatedAt:   sess.CreatedAt.Unix(),
			LastCapture: sess.LastCapture.Unix(),
		})
	}

	// Send initial session list
	jsonData, _ := json.Marshal(map[string]interface{}{
		"type":     "sessions",
		"sessions": sessionInfos,
	})
	client.Send <- jsonData

	// Start pumps
	go client.ReadPump()
	go client.WritePump()
}
