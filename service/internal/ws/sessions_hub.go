// Package ws provides WebSocket broadcasting for session updates
package ws

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// SessionHub broadcasts session updates to all connected clients
type SessionHub struct {
	clients    map[*SessionClient]bool
	broadcast  chan []byte
	register   chan *SessionClient
	unregister chan *SessionClient
	mu         sync.RWMutex
}

// SessionClient represents a WebSocket client connected for session updates
type SessionClient struct {
	Conn *websocket.Conn
	Hub  *SessionHub
	Send chan []byte
}

// NewSessionHub creates a new session hub
func NewSessionHub() *SessionHub {
	hub := &SessionHub{
		clients:    make(map[*SessionClient]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *SessionClient),
		unregister: make(chan *SessionClient),
	}
	go hub.run()
	return hub
}

// run handles hub operations
func (h *SessionHub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("Session client connected (total: %d)", len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
			h.mu.Unlock()
			log.Printf("Session client disconnected (total: %d)", len(h.clients))

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					h.mu.RUnlock()
					h.unregister <- client
					h.mu.RLock()
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Register adds a new client to the hub
func (h *SessionHub) Register(client *SessionClient) {
	h.register <- client
}

// Unregister removes a client from the hub
func (h *SessionHub) Unregister(client *SessionClient) {
	h.unregister <- client
}

// BroadcastSessions sends session list to all connected clients
func (h *SessionHub) BroadcastSessions(sessions []SessionInfo) {
	data, err := json.Marshal(map[string]interface{}{
		"type":     "sessions",
		"sessions": sessions,
	})
	if err != nil {
		log.Printf("Failed to marshal sessions: %v", err)
		return
	}

	h.broadcast <- data
}

// SessionInfo represents session information for broadcasting
type SessionInfo struct {
	ID          string `json:"id"`
	SessionName string `json:"session_name"`
	CreatedAt   int64  `json:"created_at"`
	LastCapture int64  `json:"last_capture"`
}

// ReadPump handles messages from the WebSocket connection
func (c *SessionClient) ReadPump() {
	defer func() {
		c.Hub.Unregister(c)
		_ = c.Conn.Close()
	}()

	// We don't expect any messages from clients, just keep connection alive
	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}
	}
}

// WritePump handles sending messages to the WebSocket connection
func (c *SessionClient) WritePump() {
	defer func() {
		_ = c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		}
	}
}
