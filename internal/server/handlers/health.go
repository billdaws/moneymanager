package handlers

import (
	"net/http"
	"os"

	"github.com/billdaws/moneymanager/internal/database"
	"github.com/billdaws/moneymanager/internal/kreuzberg"
)

// HealthResponse represents the health check response.
type HealthResponse struct {
	Status              string `json:"status"`
	KreuzbergAvailable  bool   `json:"kreuzberg_available"`
	GnuCashDBWritable   bool   `json:"gnucash_db_writable"`
	MetadataDBConnected bool   `json:"metadata_db_connected"`
}

// HealthHandler handles health check requests with real dependency checks.
type HealthHandler struct {
	kreuzberg   *kreuzberg.Client
	db          *database.DB
	gnucashPath string
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(kreuzbergClient *kreuzberg.Client, db *database.DB, gnucashPath string) *HealthHandler {
	return &HealthHandler{
		kreuzberg:   kreuzbergClient,
		db:          db,
		gnucashPath: gnucashPath,
	}
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	kreuzbergOK := h.kreuzberg.Health() == nil
	metadataOK := h.db.Ping() == nil
	gnucashOK := isWritable(h.gnucashPath)

	status := "healthy"
	httpStatus := http.StatusOK
	if !kreuzbergOK || !metadataOK {
		status = "degraded"
		httpStatus = http.StatusServiceUnavailable
	}

	writeJSON(w, httpStatus, HealthResponse{
		Status:              status,
		KreuzbergAvailable:  kreuzbergOK,
		GnuCashDBWritable:   gnucashOK,
		MetadataDBConnected: metadataOK,
	})
}

func isWritable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		// File doesn't exist yet — that's OK for initial setup.
		return false
	}
	// Check if it's a regular file (not a directory).
	return info.Mode().IsRegular()
}
