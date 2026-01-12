// Package main is the entry point for the Claude Web Dashboard Service
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ibrahim/remote-vibecode/internal/api"
	gottylib "github.com/ibrahim/remote-vibecode/internal/gotty"
	"github.com/ibrahim/remote-vibecode/internal/tmux"
	"github.com/ibrahim/remote-vibecode/internal/ws"
)

const (
	ServerAddress = "0.0.0.0:8080"
	Banner        = `.%%%%%...%%%%%%..%%...%%...%%%%...%%%%%%..%%%%%%.               
.%%..%%..%%......%%%.%%%..%%..%%....%%....%%.....               
.%%%%%...%%%%....%%.%.%%..%%..%%....%%....%%%%...               
.%%..%%..%%......%%...%%..%%..%%....%%....%%.....               
.%%..%%..%%%%%%..%%...%%...%%%%.....%%....%%%%%%.               
.................................................               
.%%..%%..%%%%%%..%%%%%...%%%%%%...%%%%....%%%%...%%%%%...%%%%%%.
.%%..%%....%%....%%..%%..%%......%%..%%..%%..%%..%%..%%..%%.....
.%%..%%....%%....%%%%%...%%%%....%%......%%..%%..%%..%%..%%%%...
..%%%%.....%%....%%..%%..%%......%%..%%..%%..%%..%%..%%..%%.....
...%%....%%%%%%..%%%%%...%%%%%%...%%%%....%%%%...%%%%%...%%%%%%.
................................................................`
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	fmt.Println(Banner)

	sessionHub := ws.NewSessionHub()
	tmuxMgr := tmux.New(sessionHub)
	apiHandlers := api.New()

	gottyMgr := gottylib.NewManager()
	gottyHandler := ws.NewGottyHandler(gottyMgr)
	tmuxHandlers := api.NewTmuxHandlers(tmuxMgr, sessionHub)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	router.GET("/gotty/:tmux_session", gottyHandler.HandleTmuxSession)
	router.Static("/static", "/Users/ibrahim/Lab/claude-plugin/service/web")

	apiV1 := router.Group("/api/v1")
	apiV1.GET("/tmux/sessions", tmuxHandlers.ListSessions)
	apiV1.GET("/sessions/ws", tmuxHandlers.SessionWebSocket)
	apiV1.GET("/health", apiHandlers.HealthCheck)

	router.GET("/", func(c *gin.Context) {
		c.File("/Users/ibrahim/Lab/claude-plugin/service/web/index.html")
	})
	router.NoRoute(func(c *gin.Context) {
		c.File("/Users/ibrahim/Lab/claude-plugin/service/web/index.html")
	})

	srv := &http.Server{
		Addr:    ServerAddress,
		Handler: router,
	}

	go func() {
		log.Printf("Server listening on %s", ServerAddress)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down the server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}
}
