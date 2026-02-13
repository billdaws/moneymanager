# Finance Manager Implementation Plan

## Overview
Build a Go-based finance management service that processes financial statements via Kreuzberg and loads transactions into GNU Cash.

## Architecture

```
Client → Go Service (/upload) → Kreuzberg (extract) → Normalize → GNU Cash SQLite
                              ↓
                         Metadata SQLite (tracking)
```

## Project Structure

```
moneymanager/
├── cmd/server/main.go                    # Entry point
├── internal/
│   ├── config/config.go                  # Environment-based configuration
│   ├── server/
│   │   ├── server.go                     # HTTP server setup
│   │   ├── middleware.go                 # Logging, recovery
│   │   └── handlers/
│   │       ├── upload.go                 # POST /upload endpoint
│   │       ├── health.go                 # GET /health
│   │       └── statements.go             # GET /statements
│   ├── statement/
│   │   ├── processor.go                  # Processing orchestration
│   │   ├── validator.go                  # File validation
│   │   └── store.go                      # Metadata DB access
│   ├── kreuzberg/
│   │   ├── client.go                     # Kreuzberg HTTP client
│   │   └── types.go                      # Request/response types
│   ├── transaction/
│   │   ├── normalizer.go                 # Parse & normalize
│   │   ├── categorizer.go                # Auto-categorization
│   │   └── types.go                      # Internal format
│   ├── gnucash/
│   │   ├── writer.go                     # Main writer (single-threaded queue)
│   │   ├── schema.go                     # SQL schema definitions
│   │   ├── types.go                      # Go structs for entities
│   │   ├── guid.go                       # UUID generation
│   │   ├── account.go                    # Account management
│   │   ├── transaction.go                # Transaction creation
│   │   └── split.go                      # Split balancing
│   └── database/
│       ├── metadata.go                   # Metadata SQLite
│       └── migrations.go                 # Schema migrations
├── flake.nix                              # Nix flake
├── shell.nix                              # Dev shell
├── go.mod / go.sum                        # Go dependencies
├── .env.example                           # Example config
├── docker-compose.yml                     # Kreuzberg + Caddy (existing)
├── Caddyfile                              # Reverse proxy (existing)
└── tasks/                                 # Progress tracking (markdown)
    ├── phase1-foundation.md
    ├── phase2-kreuzberg.md
    ├── phase3-gnucash-read.md
    ├── phase4-gnucash-write.md
    ├── phase5-normalization.md
    └── phase6-polish.md
```

## Critical Files

1. **cmd/server/main.go** - Wires up dependencies, starts HTTP server
2. **internal/gnucash/writer.go** - Custom GNU Cash SQLite writer (most complex)
3. **internal/server/handlers/upload.go** - Upload endpoint orchestration
4. **internal/kreuzberg/client.go** - Kreuzberg API integration
5. **internal/transaction/normalizer.go** - Transaction normalization logic

## API Design

### POST /upload
- **Input**: `multipart/form-data` with `file` field
- **Optional**: `account_type`, `account_name`, `statement_date`
- **Response**: JSON with `statement_id`, `status`, `transactions_extracted`

### GET /health
- Check server, Kreuzberg availability, database connectivity

### GET /statements
- List processed statements with transaction counts

## Metadata SQLite Schema

```sql
-- statements: Track uploaded files and processing status
CREATE TABLE statements (
    id TEXT PRIMARY KEY,              -- UUID
    filename TEXT NOT NULL,
    file_hash TEXT NOT NULL UNIQUE,   -- SHA256 (duplicate detection)
    status TEXT NOT NULL,             -- pending, processing, processed, failed
    transaction_count INTEGER,
    upload_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- transactions_raw: Store extracted transactions before GNU Cash write
CREATE TABLE transactions_raw (
    id TEXT PRIMARY KEY,
    statement_id TEXT NOT NULL,
    transaction_date DATE NOT NULL,
    description TEXT NOT NULL,
    amount_num INTEGER NOT NULL,      -- Numerator
    amount_denom INTEGER NOT NULL,    -- Denominator
    currency TEXT DEFAULT 'USD',
    category TEXT,
    gnucash_tx_guid TEXT,             -- Link to GNU Cash
    FOREIGN KEY (statement_id) REFERENCES statements(id)
);

-- processing_log: Audit trail
CREATE TABLE processing_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    statement_id TEXT NOT NULL,
    level TEXT NOT NULL,              -- info, warning, error
    stage TEXT NOT NULL,              -- upload, extraction, normalization, gnucash_write
    message TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## GNU Cash Integration

### Key Constraints
1. **GUIDs**: RFC4122 UUIDs (32 char lowercase hex, no hyphens)
2. **Balanced Transactions**: Sum of splits must equal zero
3. **Fractional Amounts**: Store as numerator/denominator (e.g., $10.50 = 1050/100)
4. **Double-Entry**: Minimum 2 splits per transaction

### Implementation Approach
1. **Read**: Open existing GNU Cash SQLite, read accounts/commodities
2. **Account Setup**: Create default structure (Assets:Bank, Expenses:*, etc.)
3. **Write**: Generate balanced splits (e.g., CreditCard -$50, Expenses:Dining +$50)
4. **Concurrency**: Single-threaded write queue (channel-based) to avoid SQLite conflicts

### GNU Cash Writer Queue
```go
type GnuCashWriter struct {
    db         *sql.DB
    writeQueue chan WriteRequest  // Single goroutine processes writes
    done       chan struct{}
}
```

## Nix Environment

### flake.nix
Provides:
- Latest Go toolchain
- gopls, gotools, golangci-lint
- Docker, docker-compose
- SQLite tools
- curl, jq

### Development Flow
```bash
nix develop              # Enter dev environment
go run cmd/server/main.go # Start server
docker-compose up -d     # Start Kreuzberg
```

## Dependencies (go.mod)

```go
require (
    github.com/mattn/go-sqlite3 v1.14.22  // SQLite driver (CGO)
    github.com/google/uuid v1.6.0          // UUID generation
)
```

Uses Go standard library for:
- net/http (server & client)
- database/sql
- encoding/json
- log/slog (structured logging)

## Configuration (.env)

```bash
SERVER_PORT=3000
KREUZBERG_URL=http://localhost:8080
GNUCASH_DB_PATH=./data/finance.gnucash
METADATA_DB_PATH=./data/metadata.db
UPLOAD_MAX_SIZE_MB=50
LOG_LEVEL=info
```

## Transaction Normalization

1. **Parse Kreuzberg Response**: Extract tables from JSON
2. **Clean**: Parse dates/amounts, trim descriptions
3. **Categorize**: Rule-based matching (e.g., "COSTCO" → "Groceries")
4. **Map to GNU Cash**: Determine source/destination accounts, create balanced splits

## Docker Setup

### Development (Recommended)
- Go app runs via Nix (outside Docker)
- Connects to Kreuzberg at `localhost:8080`
- Faster iteration, easier debugging

### Production
- Add `finance-manager` service to docker-compose.yml
- Build with Dockerfile (multi-stage: Go builder → Alpine runtime)

## Implementation Phases

### Phase 1: Foundation
- [ ] Initialize Nix flake with Go toolchain
- [ ] Create Go module (go.mod)
- [ ] Set up project directory structure
- [ ] Implement config loading (internal/config)
- [ ] Basic HTTP server with /health endpoint
- [ ] Structured logging setup

### Phase 2: Kreuzberg Integration
- [ ] Implement Kreuzberg HTTP client
- [ ] File upload handler (multipart/form-data)
- [ ] Parse Kreuzberg JSON response
- [ ] Metadata SQLite schema & migrations
- [ ] Store uploaded statements and raw transactions

### Phase 3: GNU Cash Library - Read Operations
- [ ] Open GNU Cash SQLite database
- [ ] Read books, commodities, accounts
- [ ] Implement GUID generation (RFC4122 format)
- [ ] Map account names to GUIDs
- [ ] Understand account hierarchy

### Phase 4: GNU Cash Library - Write Operations
- [ ] Create default account structure
- [ ] Implement transaction writer
- [ ] Implement split creation & balancing
- [ ] Write single-threaded queue for concurrency
- [ ] Test with real GNU Cash application

### Phase 5: Transaction Normalization
- [ ] Parse Kreuzberg tables (dates, amounts, descriptions)
- [ ] Normalize to internal format
- [ ] Implement basic categorization rules
- [ ] End-to-end upload → GNU Cash flow
- [ ] GET /statements endpoint

### Phase 6: Polish
- [ ] Error handling & recovery
- [ ] Duplicate detection (SHA256 hash)
- [ ] Statement detail endpoint (GET /statements/:id)
- [ ] Docker production packaging
- [ ] README documentation

## Critical Design Decisions

1. **Standard Library HTTP**: No framework (per requirements), more control
2. **Custom GNU Cash Library**: Full control, no licensing issues (per requirements)
3. **SQLite for Metadata**: Embedded, matches GNU Cash storage
4. **Single-Threaded GNU Cash Writer**: Avoids SQLite write conflicts
5. **Separate Metadata DB**: Don't pollute GNU Cash database

## Verification

### End-to-End Test
1. Start Kreuzberg: `docker-compose up -d`
2. Start app: `air` or `go run cmd/server/main.go`
3. Upload test statement: `curl -F "file=@test.pdf" http://localhost:3000/upload`
4. Check response for `transactions_extracted` count
5. Open GNU Cash application with `./data/finance.gnucash`
6. Verify transactions appear with correct amounts and accounts

### Unit Tests
- GUID generation format
- Decimal arithmetic (numerator/denominator)
- Category pattern matching
- Date parsing (multiple formats)

### Integration Tests
- Kreuzberg client with mock server
- GNU Cash writer with test database
- Full upload flow with sample statements

## Security Considerations

- File size limits (50MB default)
- MIME type validation
- SQL injection prevention (parameterized queries)
- Path traversal prevention
- Kreuzberg isolated on internal Docker network (no egress)

## Progress Tracking

Create markdown files in `tasks/` directory for each phase:
- Track completion status (checkboxes)
- Document blockers and decisions
- Note testing results
- Log questions and resolutions
