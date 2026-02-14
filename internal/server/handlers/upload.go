package handlers

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/billdaws/moneymanager/internal/statement"
)

// UploadHandler handles POST /upload requests.
type UploadHandler struct {
	processor  *statement.Processor
	maxSizeMB int
	logger     *slog.Logger
}

// NewUploadHandler creates a new UploadHandler.
func NewUploadHandler(processor *statement.Processor, maxSizeMB int, logger *slog.Logger) *UploadHandler {
	return &UploadHandler{
		processor: processor,
		maxSizeMB: maxSizeMB,
		logger:    logger,
	}
}

type uploadResponse struct {
	StatementID           string `json:"statement_id"`
	Filename              string `json:"filename"`
	Status                string `json:"status"`
	TransactionsExtracted int    `json:"transactions_extracted"`
	ProcessingTimeMs      int64  `json:"processing_time_ms"`
	Duplicate             bool   `json:"duplicate"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (h *UploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limit the request body to maxSizeMB + 1MB overhead for form fields.
	maxBytes := int64(h.maxSizeMB+1) * 1024 * 1024
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

	if err := r.ParseMultipartForm(maxBytes); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "failed to parse multipart form: " + err.Error()})
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "missing or invalid 'file' field"})
		return
	}
	defer func() { _ = file.Close() }()

	data, err := io.ReadAll(file)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "failed to read file: " + err.Error()})
		return
	}

	accountType := r.FormValue("account_type")
	accountName := r.FormValue("account_name")
	statementDate := r.FormValue("statement_date")

	result, err := h.processor.Process(header.Filename, data, accountType, accountName, statementDate)
	if err != nil {
		h.logger.Error("processing failed",
			"filename", header.Filename,
			"error", err,
		)
		writeJSON(w, http.StatusUnprocessableEntity, errorResponse{Error: err.Error()})
		return
	}

	status := http.StatusOK
	if result.Duplicate {
		status = http.StatusOK
	}

	writeJSON(w, status, uploadResponse{
		StatementID:           result.StatementID,
		Filename:              result.Filename,
		Status:                result.Status,
		TransactionsExtracted: result.TransactionsExtracted,
		ProcessingTimeMs:      result.ProcessingTimeMs,
		Duplicate:             result.Duplicate,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
