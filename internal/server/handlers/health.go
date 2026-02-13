package handlers

import (
	"encoding/json"
	"net/http"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status              string `json:"status"`
	KreuzbergAvailable  bool   `json:"kreuzberg_available"`
	GnuCashDBWritable   bool   `json:"gnucash_db_writable"`
	MetadataDBConnected bool   `json:"metadata_db_connected"`
}

// HealthHandler handles health check requests
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Add actual health checks for:
	// - Kreuzberg connectivity
	// - GNU Cash database writability
	// - Metadata database connectivity

	response := HealthResponse{
		Status:              "healthy",
		KreuzbergAvailable:  true,  // TODO: Actual check
		GnuCashDBWritable:   true,  // TODO: Actual check
		MetadataDBConnected: true,  // TODO: Actual check
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
