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
	"github.com/ibrahim/remote-vibecode/cmd/vibecode/commands"
	"github.com/ibrahim/remote-vibecode/internal/api"
	gottylib "github.com/ibrahim/remote-vibecode/internal/gotty"
	"github.com/ibrahim/remote-vibecode/internal/tmux"
	"github.com/ibrahim/remote-vibecode/internal/ws"
	"github.com/spf13/cobra"
)

//go:embed web
var webFS embed.FS

const (
	DefaultHost = "127.0.0.1"
	DefaultPort = "7676"
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

var (
	serveHost string
	servePort string
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the rv web server",
	Long:  "Start the web server for remote terminal viewing. Use --host and --port to customize.",
	RunE:  runServe,
}

func init() {
	serveCmd.Flags().StringVar(&serveHost, "host", DefaultHost, "Host to bind to")
	serveCmd.Flags().StringVar(&servePort, "port", DefaultPort, "Port to listen on")
}

func runServe(cmd *cobra.Command, args []string) error {
	gin.SetMode(gin.ReleaseMode)

	serverAddr := fmt.Sprintf("%s:%s", serveHost, servePort)
	fmt.Printf(Banner, serverAddr, serverAddr)

	sessionHub := ws.NewSessionHub()
	tmuxMgr := tmux.New(sessionHub)
	apiHandlers := api.New()

	gottyMgr := gottylib.NewManager()
	gottyHandler := ws.NewGottyHandler(gottyMgr)
	tmuxHandlers := api.NewTmuxHandlers(tmuxMgr, sessionHub)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	router.RedirectTrailingSlash = false
	router.RedirectFixedPath = false

	webSubFS, err := fs.Sub(webFS, "web")
	if err != nil {
		return fmt.Errorf("failed to get web subdirectory: %w", err)
	}

	router.GET("/static/*filepath", func(c *gin.Context) {
		filepath := c.Param("filepath")
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

	router.GET("/", func(c *gin.Context) {
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

	return nil
}

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

func main() {
	rootCmd := &cobra.Command{
		Use:   "rv",
		Short: "rv - Remote tmux session management",
		Long: `rv is a CLI tool for managing tmux sessions used with
the remote vibecode service. It provides an easy way to create, join,
list, and stop tmux sessions with custom configuration.`,
		Version: "1.0.0",
	}

	// Add subcommands
	rootCmd.AddCommand(commands.StartCmd)
	rootCmd.AddCommand(commands.JoinCmd)
	rootCmd.AddCommand(commands.ListCmd)
	rootCmd.AddCommand(commands.StopCmd)
	rootCmd.AddCommand(serveCmd)

	// Run the command
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
