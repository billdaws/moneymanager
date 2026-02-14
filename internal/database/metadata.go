package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

// DB wraps a SQLite connection for the metadata database.
type DB struct {
	conn *sql.DB
}

// Statement represents a row in the statements table.
type Statement struct {
	ID               string
	Filename         string
	FileHash         string
	FileSize         int64
	MimeType         string
	Status           string
	TransactionCount int
	AccountType      string
	AccountName      string
	StatementDate    string
	ErrorMessage     string
	UploadTime       time.Time
	ProcessedTime    time.Time
}

// TransactionRaw represents a row in the transactions_raw table.
type TransactionRaw struct {
	ID          string
	StatementID string
	RowIndex    int
	Headers     string // JSON array
	RawData     string // JSON array
	CreatedAt   time.Time
}

// LogEntry represents a row in the processing_log table.
type LogEntry struct {
	ID          int64
	StatementID string
	Level       string
	Stage       string
	Message     string
	CreatedAt   time.Time
}

// Open creates a connection to the metadata SQLite database and runs migrations.
func Open(dbPath string) (*DB, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}

	conn, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_foreign_keys=ON")
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	if _, err := conn.Exec(schema); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	return &DB{conn: conn}, nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.conn.Close()
}

// Ping checks that the database is reachable.
func (db *DB) Ping() error {
	return db.conn.Ping()
}

// CreateStatement inserts a new statement record and returns its ID.
func (db *DB) CreateStatement(filename, fileHash string, fileSize int64, mimeType, accountType, accountName, statementDate string) (string, error) {
	id := uuid.New().String()
	now := time.Now().UTC().Format(time.RFC3339)

	_, err := db.conn.Exec(`
		INSERT INTO statements (id, filename, file_hash, file_size, mime_type, status, account_type, account_name, statement_date, upload_time)
		VALUES (?, ?, ?, ?, ?, 'pending', ?, ?, ?, ?)`,
		id, filename, fileHash, fileSize, mimeType, accountType, accountName, statementDate, now,
	)
	if err != nil {
		return "", fmt.Errorf("insert statement: %w", err)
	}

	return id, nil
}

// GetStatementByHash returns a statement by its file hash, or nil if not found.
func (db *DB) GetStatementByHash(fileHash string) (*Statement, error) {
	row := db.conn.QueryRow(`
		SELECT id, filename, file_hash, file_size, mime_type, status, transaction_count,
		       account_type, account_name, statement_date, error_message, upload_time, processed_time
		FROM statements WHERE file_hash = ?`, fileHash)

	return scanStatement(row)
}

// GetStatement returns a statement by its ID, or nil if not found.
func (db *DB) GetStatement(id string) (*Statement, error) {
	row := db.conn.QueryRow(`
		SELECT id, filename, file_hash, file_size, mime_type, status, transaction_count,
		       account_type, account_name, statement_date, error_message, upload_time, processed_time
		FROM statements WHERE id = ?`, id)

	return scanStatement(row)
}

// UpdateStatus sets the status of a statement.
func (db *DB) UpdateStatus(id, status string) error {
	_, err := db.conn.Exec(`UPDATE statements SET status = ? WHERE id = ?`, status, id)
	return err
}

// MarkProcessed marks a statement as processed with a transaction count.
func (db *DB) MarkProcessed(id string, transactionCount int) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := db.conn.Exec(`
		UPDATE statements SET status = 'processed', transaction_count = ?, processed_time = ? WHERE id = ?`,
		transactionCount, now, id,
	)
	return err
}

// MarkFailed marks a statement as failed with an error message.
func (db *DB) MarkFailed(id, errorMessage string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := db.conn.Exec(`
		UPDATE statements SET status = 'failed', error_message = ?, processed_time = ? WHERE id = ?`,
		errorMessage, now, id,
	)
	return err
}

// InsertTransactionRaw inserts a raw transaction row.
func (db *DB) InsertTransactionRaw(statementID string, rowIndex int, headers, rawData string) (string, error) {
	id := uuid.New().String()
	now := time.Now().UTC().Format(time.RFC3339)

	_, err := db.conn.Exec(`
		INSERT INTO transactions_raw (id, statement_id, row_index, headers, raw_data, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		id, statementID, rowIndex, headers, rawData, now,
	)
	if err != nil {
		return "", fmt.Errorf("insert transaction_raw: %w", err)
	}

	return id, nil
}

// InsertLogEntry inserts a processing log entry.
func (db *DB) InsertLogEntry(statementID, level, stage, message string) error {
	now := time.Now().UTC().Format(time.RFC3339)

	_, err := db.conn.Exec(`
		INSERT INTO processing_log (statement_id, level, stage, message, created_at)
		VALUES (?, ?, ?, ?, ?)`,
		statementID, level, stage, message, now,
	)
	return err
}

func scanStatement(row *sql.Row) (*Statement, error) {
	var s Statement
	var uploadTime, processedTime string

	err := row.Scan(
		&s.ID, &s.Filename, &s.FileHash, &s.FileSize, &s.MimeType,
		&s.Status, &s.TransactionCount,
		&s.AccountType, &s.AccountName, &s.StatementDate,
		&s.ErrorMessage, &uploadTime, &processedTime,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan statement: %w", err)
	}

	if t, err := time.Parse(time.RFC3339, uploadTime); err == nil {
		s.UploadTime = t
	}
	if t, err := time.Parse(time.RFC3339, processedTime); err == nil {
		s.ProcessedTime = t
	}

	return &s, nil
}
