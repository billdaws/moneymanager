package statement

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/billdaws/moneymanager/internal/kreuzberg"
)

// ProcessResult contains the outcome of processing a statement upload.
type ProcessResult struct {
	StatementID           string
	Filename              string
	Status                string
	TransactionsExtracted int
	ProcessingTimeMs      int64
	Duplicate             bool
}

// Processor orchestrates statement processing: validate → hash → dedup → extract → store.
type Processor struct {
	store        *Store
	kreuzberg    *kreuzberg.Client
	maxSizeMB    int
	allowedTypes []string
	logger       *slog.Logger
}

// NewProcessor creates a new Processor.
func NewProcessor(store *Store, kreuzbergClient *kreuzberg.Client, maxSizeMB int, allowedTypes []string, logger *slog.Logger) *Processor {
	return &Processor{
		store:        store,
		kreuzberg:    kreuzbergClient,
		maxSizeMB:    maxSizeMB,
		allowedTypes: allowedTypes,
		logger:       logger,
	}
}

// Process handles the full lifecycle of a statement upload.
func (p *Processor) Process(filename string, data []byte, accountType, accountName, statementDate string) (*ProcessResult, error) {
	start := time.Now()

	// 1. Validate file type and size.
	mimeType, err := ValidateFile(data, p.maxSizeMB, p.allowedTypes)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// 2. Compute SHA256 hash.
	fileHash := HashFile(data)

	// 3. Check for duplicate.
	existing, err := p.store.FindDuplicate(fileHash)
	if err != nil {
		return nil, fmt.Errorf("duplicate check: %w", err)
	}
	if existing != nil {
		return &ProcessResult{
			StatementID:           existing.ID,
			Filename:              existing.Filename,
			Status:                existing.Status,
			TransactionsExtracted: existing.TransactionCount,
			ProcessingTimeMs:      time.Since(start).Milliseconds(),
			Duplicate:             true,
		}, nil
	}

	// 4. Create statement record.
	statementID, err := p.store.CreateStatement(filename, fileHash, int64(len(data)), mimeType, accountType, accountName, statementDate)
	if err != nil {
		return nil, fmt.Errorf("create statement: %w", err)
	}

	p.store.Log(statementID, "info", "upload", "Statement created")

	// 5. Mark as processing.
	if err := p.store.MarkProcessing(statementID); err != nil {
		return nil, fmt.Errorf("mark processing: %w", err)
	}

	// 6. Send to Kreuzberg for extraction.
	p.store.Log(statementID, "info", "extraction", "Sending to Kreuzberg")

	results, err := p.kreuzberg.Extract(filename, data, mimeType)
	if err != nil {
		p.store.Log(statementID, "error", "extraction", err.Error())
		_ = p.store.MarkFailed(statementID, err.Error())

		p.logger.Error("kreuzberg extraction failed",
			"statement_id", statementID,
			"error", err,
		)

		return &ProcessResult{
			StatementID:      statementID,
			Filename:         filename,
			Status:           "failed",
			ProcessingTimeMs: time.Since(start).Milliseconds(),
		}, nil
	}

	p.store.Log(statementID, "info", "extraction", fmt.Sprintf("Received %d extraction results", len(results)))

	// 7. Store table rows as raw transactions.
	rowCount, err := p.store.StoreExtractionResults(statementID, results)
	if err != nil {
		p.store.Log(statementID, "error", "storage", err.Error())
		_ = p.store.MarkFailed(statementID, err.Error())

		return &ProcessResult{
			StatementID:      statementID,
			Filename:         filename,
			Status:           "failed",
			ProcessingTimeMs: time.Since(start).Milliseconds(),
		}, nil
	}

	// 8. Mark as processed.
	if err := p.store.MarkProcessed(statementID, rowCount); err != nil {
		return nil, fmt.Errorf("mark processed: %w", err)
	}

	p.store.Log(statementID, "info", "complete", fmt.Sprintf("Processed %d transactions", rowCount))

	p.logger.Info("statement processed",
		"statement_id", statementID,
		"filename", filename,
		"transactions", rowCount,
		"duration_ms", time.Since(start).Milliseconds(),
	)

	return &ProcessResult{
		StatementID:           statementID,
		Filename:              filename,
		Status:                "processed",
		TransactionsExtracted: rowCount,
		ProcessingTimeMs:      time.Since(start).Milliseconds(),
	}, nil
}
