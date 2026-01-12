// Package main is the entry point for the remote-vibecode service
package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ibrahim/remote-vibecode/internal/api"
	gottylib "github.com/ibrahim/remote-vibecode/internal/gotty"
	"github.com/ibrahim/remote-vibecode/internal/tmux"
	"github.com/ibrahim/remote-vibecode/internal/ws"
)

//go:embed web/*
var webFS embed.FS

var Version = "dev"

const (
	DefaultHost = "0.0.0.0"
	DefaultPort = "8080"
	Banner      = `.................................................
.#####...######..##...##...####...######..######.               
.##..##..##......###.###..##..##....##....##.....               
.#####...####....##.#.##..##..##....##....####...               
.##..##..##......##...##..##..##....##....##.....               
.##..##..######..##...##...####.....##....######.               
.................................................               
.##..##..######..#####...######...####....####...#####...######.
.##..##....##....##..##..##......##..##..##..##..##..##..##.....
.##..##....##....#####...####....##......##..##..##..##..####...
..####.....##....##..##..##......##..##..##..##..##..##..##.....
...##....######..#####...######...####....####...#####...######.
................................................................
Remote vibecode service listening on %s
Web UI available at http://%s
.................................................
`
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	// Get configuration from environment
	host := getEnv("HOST", DefaultHost)
	port := getEnv("PORT", DefaultPort)
	serverAddr := fmt.Sprintf("%s:%s", host, port)

	// Print banner with server address
	fmt.Printf(Banner, serverAddr, serverAddr)
	if Version != "dev" {
		fmt.Printf("remote-vibecode version %s\n", Version)
	}

	sessionHub := ws.NewSessionHub()
	tmuxMgr := tmux.New(sessionHub)
	apiHandlers := api.New()

	gottyMgr := gottylib.NewManager()
	gottyHandler := ws.NewGottyHandler(gottyMgr)
	tmuxHandlers := api.NewTmuxHandlers(tmuxMgr, sessionHub)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// Disable automatic redirects that can cause loops
	router.RedirectTrailingSlash = false
	router.RedirectFixedPath = false

	// Extract web subdirectory from embedded FS
	webSubFS, err := fs.Sub(webFS, "web")
	if err != nil {
		log.Fatalf("Failed to get web subdirectory: %v", err)
	}

	// Serve static files from embedded filesystem
	// Strip /static/ prefix since files are in web/ root, not web/static/
	router.GET("/static/*filepath", func(c *gin.Context) {
		filepath := c.Param("filepath")
		// Trim leading slash - fs.ReadFile expects relative paths
		if strings.HasPrefix(filepath, "/") {
			filepath = filepath[1:]
		}
		content, err := fs.ReadFile(webSubFS, filepath)
		if err != nil {
			log.Printf("Error reading static file %s: %v", filepath, err)
			c.Status(404)
			return
		}
		c.Header("Content-Type", getContentType(filepath))
		c.Data(200, "", content)
	})

	// Serve index.html at root
	router.GET("/", func(c *gin.Context) {
		// Try without leading slash first
		content, err := fs.ReadFile(webSubFS, "index.html")
		if err != nil {
			log.Printf("Error reading index.html: %v", err)
			c.Status(404)
			return
		}
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.Data(200, "", content)
	})

	router.GET("/gotty/:tmux_session", gottyHandler.HandleTmuxSession)

	apiV1 := router.Group("/api/v1")
	apiV1.GET("/tmux/sessions", tmuxHandlers.ListSessions)
	apiV1.GET("/sessions/ws", tmuxHandlers.SessionWebSocket)
	apiV1.GET("/health", apiHandlers.HealthCheck)

	srv := &http.Server{
		Addr:    serverAddr,
		Handler: router,
	}

	go func() {
		log.Printf("Server started on %s", serverAddr)
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

// getEnv retrieves environment variable or returns default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getContentType returns the content type for a file based on its extension
func getContentType(filepath string) string {
	switch {
	case strings.HasSuffix(filepath, ".html"):
		return "text/html; charset=utf-8"
	case strings.HasSuffix(filepath, ".css"):
		return "text/css; charset=utf-8"
	case strings.HasSuffix(filepath, ".js"):
		return "application/javascript; charset=utf-8"
	case strings.HasSuffix(filepath, ".json"):
		return "application/json"
	case strings.HasSuffix(filepath, ".png"):
		return "image/png"
	case strings.HasSuffix(filepath, ".jpg"), strings.HasSuffix(filepath, ".jpeg"):
		return "image/jpeg"
	case strings.HasSuffix(filepath, ".svg"):
		return "image/svg+xml"
	case strings.HasSuffix(filepath, ".ico"):
		return "image/x-icon"
	default:
		return "application/octet-stream"
	}
}
