package database

const schema = `
PRAGMA journal_mode=WAL;
PRAGMA foreign_keys=ON;

CREATE TABLE IF NOT EXISTS statements (
	id              TEXT PRIMARY KEY,
	filename        TEXT NOT NULL,
	file_hash       TEXT NOT NULL UNIQUE,
	file_size       INTEGER NOT NULL,
	mime_type       TEXT NOT NULL,
	status          TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','processing','processed','failed')),
	transaction_count INTEGER NOT NULL DEFAULT 0,
	account_type    TEXT NOT NULL DEFAULT '',
	account_name    TEXT NOT NULL DEFAULT '',
	statement_date  TEXT NOT NULL DEFAULT '',
	error_message   TEXT NOT NULL DEFAULT '',
	upload_time     TEXT NOT NULL,
	processed_time  TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_statements_file_hash ON statements(file_hash);
CREATE INDEX IF NOT EXISTS idx_statements_status ON statements(status);

CREATE TABLE IF NOT EXISTS transactions_raw (
	id           TEXT PRIMARY KEY,
	statement_id TEXT NOT NULL,
	row_index    INTEGER NOT NULL,
	headers      TEXT NOT NULL DEFAULT '[]',
	raw_data     TEXT NOT NULL DEFAULT '[]',
	created_at   TEXT NOT NULL,
	FOREIGN KEY (statement_id) REFERENCES statements(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_transactions_raw_statement_id ON transactions_raw(statement_id);

CREATE TABLE IF NOT EXISTS processing_log (
	id           INTEGER PRIMARY KEY AUTOINCREMENT,
	statement_id TEXT NOT NULL,
	level        TEXT NOT NULL,
	stage        TEXT NOT NULL,
	message      TEXT NOT NULL,
	created_at   TEXT NOT NULL,
	FOREIGN KEY (statement_id) REFERENCES statements(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_processing_log_statement_id ON processing_log(statement_id);
`
