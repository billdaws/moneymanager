package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/billdaws/moneymanager/internal/config"
	"github.com/billdaws/moneymanager/internal/server/handlers"
)

// Server wraps the HTTP server
type Server struct {
	httpServer *http.Server
	logger     *slog.Logger
}

// New creates a new HTTP server
func New(cfg *config.Config, logger *slog.Logger) *Server {
	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("/health", handlers.HealthHandler)

	// Apply middleware
	handler := CORSMiddleware(mux)
	handler = LoggingMiddleware(logger)(handler)
	handler = RecoveryMiddleware(logger)(handler)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      handler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	return &Server{
		httpServer: httpServer,
		logger:     logger,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.logger.Info("starting http server",
		"addr", s.httpServer.Addr,
	)
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down http server")
	return s.httpServer.Shutdown(ctx)
}

// Addr returns the server address
func (s *Server) Addr() string {
	return s.httpServer.Addr
}
