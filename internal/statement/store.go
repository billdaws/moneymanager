package statement

import (
	"encoding/json"
	"fmt"

	"github.com/billdaws/moneymanager/internal/database"
	"github.com/billdaws/moneymanager/internal/kreuzberg"
)

// Store wraps DB operations for the statement domain.
type Store struct {
	db *database.DB
}

// NewStore creates a new Store.
func NewStore(db *database.DB) *Store {
	return &Store{db: db}
}

// FindDuplicate checks if a file with the same hash already exists.
// Returns the existing statement or nil.
func (s *Store) FindDuplicate(fileHash string) (*database.Statement, error) {
	return s.db.GetStatementByHash(fileHash)
}

// CreateStatement creates a new statement record.
func (s *Store) CreateStatement(filename, fileHash string, fileSize int64, mimeType, accountType, accountName, statementDate string) (string, error) {
	return s.db.CreateStatement(filename, fileHash, fileSize, mimeType, accountType, accountName, statementDate)
}

// MarkProcessing sets the statement status to "processing".
func (s *Store) MarkProcessing(id string) error {
	return s.db.UpdateStatus(id, "processing")
}

// StoreExtractionResults stores the table rows from a Kreuzberg extraction as raw transactions.
// Returns the total number of rows stored.
func (s *Store) StoreExtractionResults(statementID string, results []kreuzberg.ExtractionResult) (int, error) {
	totalRows := 0

	for _, result := range results {
		for _, table := range result.Tables {
			headersJSON, err := json.Marshal(table.Headers)
			if err != nil {
				return totalRows, fmt.Errorf("marshal headers: %w", err)
			}

			for _, row := range table.Rows {
				rowJSON, err := json.Marshal(row)
				if err != nil {
					return totalRows, fmt.Errorf("marshal row: %w", err)
				}

				if _, err := s.db.InsertTransactionRaw(statementID, totalRows, string(headersJSON), string(rowJSON)); err != nil {
					return totalRows, fmt.Errorf("insert row %d: %w", totalRows, err)
				}
				totalRows++
			}
		}
	}

	return totalRows, nil
}

// MarkProcessed marks a statement as processed with a transaction count.
func (s *Store) MarkProcessed(id string, transactionCount int) error {
	return s.db.MarkProcessed(id, transactionCount)
}

// MarkFailed marks a statement as failed with an error message.
func (s *Store) MarkFailed(id, errorMessage string) error {
	return s.db.MarkFailed(id, errorMessage)
}

// Log writes a processing log entry.
func (s *Store) Log(statementID, level, stage, message string) {
	// Best-effort logging; errors are silently ignored.
	_ = s.db.InsertLogEntry(statementID, level, stage, message)
}
