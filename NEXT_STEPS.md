# Next Steps: Phase 2 - Kreuzberg Integration

## Quick Start Commands

```bash
# Enter development environment
nix develop

# Start Kreuzberg
docker-compose up -d

# Run the application
go run cmd/server/main.go

# In another terminal, test health
curl http://localhost:3000/health
```

## Phase 2 Implementation Order

### Step 1: Add Dependencies

```bash
go get github.com/mattn/go-sqlite3
go get github.com/google/uuid
```

### Step 2: Implement Metadata Database

**File: `internal/database/migrations.go`**
- Define SQL schema for statements, transactions_raw, processing_log tables
- Implement InitSchema() function to create tables if they don't exist
- Add schema version tracking

**File: `internal/database/metadata.go`**
- Open SQLite connection
- CRUD operations for statements table
- CRUD operations for transactions_raw table
- Insert operations for processing_log

### Step 3: Implement Kreuzberg Client

**File: `internal/kreuzberg/types.go`**
```go
type ExtractionResult struct {
    Content           string                 `json:"content"`
    MimeType          string                 `json:"mime_type"`
    Metadata          map[string]interface{} `json:"metadata"`
    Tables            []Table                `json:"tables"`
    DetectedLanguages []string               `json:"detected_languages"`
}

type Table struct {
    Headers []string   `json:"headers"`
    Rows    [][]string `json:"rows"`
}
```

**File: `internal/kreuzberg/client.go`**
```go
type Client struct {
    baseURL    string
    httpClient *http.Client
}

func NewClient(baseURL string) *Client
func (c *Client) ExtractTransactions(file io.Reader, filename string) (*ExtractionResult, error)
```

Implementation:
1. Create multipart form with file
2. POST to baseURL + "/extract" (or check Kreuzberg docs for correct endpoint)
3. Parse JSON response
4. Return ExtractionResult or error

### Step 4: Implement Statement Processing

**File: `internal/statement/validator.go`**
```go
func ValidateFile(filename string, size int64, mimeType string) error
func HashFile(r io.Reader) (string, error)  // SHA256
```

**File: `internal/statement/store.go`**
- Wrapper around database operations specific to statements
- CheckDuplicate(hash string) (bool, error)
- Create(statement Statement) error
- UpdateStatus(id string, status string) error

**File: `internal/statement/processor.go`**
```go
type Processor struct {
    kreuzbergClient *kreuzberg.Client
    store          *Store
    logger         *slog.Logger
}

func (p *Processor) ProcessUpload(file io.Reader, filename string, metadata UploadMetadata) (*ProcessResult, error)
```

Flow:
1. Validate file
2. Hash file for duplicate detection
3. Check if duplicate exists
4. Create statement record (status: processing)
5. Call Kreuzberg
6. Store raw transactions
7. Update statement (status: processed)
8. Return result

### Step 5: Implement Upload Handler

**File: `internal/server/handlers/upload.go`**
```go
type UploadHandler struct {
    processor *statement.Processor
    logger    *slog.Logger
    maxSize   int64
}

func NewUploadHandler(processor *statement.Processor, maxSizeMB int, logger *slog.Logger) *UploadHandler

func (h *UploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request)
```

Implementation:
1. Check method is POST
2. Parse multipart form (max 50MB or configured size)
3. Extract file and optional metadata (account_type, account_name, statement_date)
4. Call processor.ProcessUpload()
5. Return JSON response

Response format:
```json
{
  "statement_id": "uuid",
  "filename": "statement.pdf",
  "status": "processed",
  "transactions_extracted": 42,
  "processing_time_ms": 1523,
  "errors": []
}
```

### Step 6: Wire Up in Server

**Update: `internal/server/server.go`**
```go
func New(cfg *config.Config, logger *slog.Logger) *Server {
    // Initialize database
    db := database.NewMetadataDB(cfg.Database.MetadataPath)
    db.InitSchema()

    // Initialize Kreuzberg client
    kreuzbergClient := kreuzberg.NewClient(cfg.Kreuzberg.URL)

    // Initialize processor
    processor := statement.NewProcessor(kreuzbergClient, db, logger)

    // Create handlers
    uploadHandler := handlers.NewUploadHandler(processor, cfg.Upload.MaxSizeMB, logger)

    mux := http.NewServeMux()
    mux.HandleFunc("/health", handlers.HealthHandler)
    mux.Handle("/upload", uploadHandler)

    // ... rest of setup
}
```

### Step 7: Update Health Check

**Update: `internal/server/handlers/health.go`**

Add actual checks for:
- Kreuzberg connectivity (simple HTTP GET to /health or /)
- Metadata database connectivity (simple query)
- GNU Cash database writability (check file exists and is writable)

## Testing Phase 2

### Test 1: Database Schema
```bash
go run cmd/server/main.go &
sleep 2
sqlite3 data/metadata.db ".schema"
# Should show statements, transactions_raw, processing_log tables
pkill -f "cmd/server"
```

### Test 2: Kreuzberg Connectivity
```bash
# Start Kreuzberg
docker-compose up -d

# Check if Kreuzberg is responding
curl http://localhost:8080/

# Start our service
go run cmd/server/main.go &

# Test health - should show Kreuzberg available
curl http://localhost:3000/health
```

### Test 3: File Upload
```bash
# Create a test PDF or use existing statement
curl -F "file=@test.pdf" -F "account_type=credit_card" http://localhost:3000/upload

# Check database
sqlite3 data/metadata.db "SELECT * FROM statements;"
sqlite3 data/metadata.db "SELECT COUNT(*) FROM transactions_raw;"
```

### Test 4: Duplicate Detection
```bash
# Upload same file twice
curl -F "file=@test.pdf" http://localhost:3000/upload
curl -F "file=@test.pdf" http://localhost:3000/upload

# Second upload should be rejected or flagged as duplicate
```

## Key Files to Create

```
internal/
├── database/
│   ├── metadata.go      # Database connection and operations
│   └── migrations.go    # Schema definitions and initialization
├── kreuzberg/
│   ├── client.go        # HTTP client for Kreuzberg API
│   └── types.go         # Request/response types
├── statement/
│   ├── processor.go     # Main processing orchestration
│   ├── validator.go     # File validation and hashing
│   └── store.go         # Database access for statements
└── server/handlers/
    └── upload.go        # POST /upload endpoint
```

## Kreuzberg API Investigation

Before implementing, check Kreuzberg documentation:
```bash
docker-compose up -d
curl http://localhost:8080/docs  # or check other common doc endpoints
curl http://localhost:8080/      # might return API info
```

Look for:
- Correct endpoint path (e.g., /extract, /parse, /api/extract)
- Required headers
- Expected form field names
- Response format

## Error Handling Considerations

1. **Network Errors**: Kreuzberg unreachable
2. **File Errors**: Invalid format, corrupted file
3. **Database Errors**: SQLite locked, disk full
4. **Duplicate Files**: Same hash already processed
5. **Parsing Errors**: Kreuzberg returns unexpected format

For each error, update statement status to "failed" with error message in processing_log.

## Success Criteria for Phase 2

- [ ] Metadata database schema created automatically
- [ ] Can upload PDF/CSV file via POST /upload
- [ ] File is sent to Kreuzberg and response is received
- [ ] Statement record is created in database with UUID
- [ ] Raw transactions are stored in transactions_raw table
- [ ] Duplicate files are detected (same SHA256 hash)
- [ ] Health endpoint shows accurate Kreuzberg connectivity
- [ ] Upload endpoint returns proper JSON response
- [ ] Processing is logged to processing_log table
- [ ] Error cases are handled gracefully

## After Phase 2

Phase 3 will implement GNU Cash read operations:
- Open existing GNU Cash SQLite database
- Read accounts, commodities, books
- Implement GUID generation
- Map account hierarchies

This is separate from Phase 2 and can wait until statement upload is working end-to-end.
