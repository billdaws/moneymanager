package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/billdaws/moneymanager/internal/config"
	"github.com/billdaws/moneymanager/internal/database"
	"github.com/billdaws/moneymanager/internal/kreuzberg"
	"github.com/billdaws/moneymanager/internal/server/handlers"
	"github.com/billdaws/moneymanager/internal/statement"
)

// Server wraps the HTTP server and its dependencies.
type Server struct {
	httpServer *http.Server
	db         *database.DB
	logger     *slog.Logger
}

// New creates a new HTTP server with all dependencies initialized.
func New(cfg *config.Config, logger *slog.Logger) (*Server, error) {
	// Open metadata database (creates file and runs migrations).
	db, err := database.Open(cfg.Database.MetadataPath)
	if err != nil {
		return nil, fmt.Errorf("open metadata database: %w", err)
	}

	// Create Kreuzberg client.
	kreuzbergClient := kreuzberg.NewClient(cfg.Kreuzberg.URL, cfg.Kreuzberg.Timeout)

	// Create statement processing pipeline.
	store := statement.NewStore(db)
	processor := statement.NewProcessor(store, kreuzbergClient, cfg.Upload.MaxSizeMB, cfg.Upload.AllowedTypes, logger)

	// Create handlers.
	healthHandler := handlers.NewHealthHandler(kreuzbergClient, db, cfg.Database.GnuCashPath)
	uploadHandler := handlers.NewUploadHandler(processor, cfg.Upload.MaxSizeMB, logger)

	// Register routes.
	mux := http.NewServeMux()
	mux.Handle("/health", healthHandler)
	mux.Handle("/upload", uploadHandler)

	// Apply middleware.
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
		db:         db,
		logger:     logger,
	}, nil
}

// Start starts the HTTP server.
func (s *Server) Start() error {
	s.logger.Info("starting http server",
		"addr", s.httpServer.Addr,
	)
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server and closes the database.
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down http server")

	err := s.httpServer.Shutdown(ctx)

	if dbErr := s.db.Close(); dbErr != nil {
		s.logger.Error("failed to close database", "error", dbErr)
	}

	return err
}

// Addr returns the server address.
func (s *Server) Addr() string {
	return s.httpServer.Addr
}
