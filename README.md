# Money Manager

A Go-based finance management service that processes financial statements and loads transactions into GNU Cash.

## Features

- Accept financial statements via `/upload` API (PDF, CSV, etc.)
- Extract transactions using Kreuzberg service
- Normalize transactions to standard format
- Load transactions into GNU Cash database
- Custom GNU Cash SQLite library implementation

## Architecture

```
Client → Go Service (/upload) → Kreuzberg (extract) → Normalize → GNU Cash SQLite
                              ↓
                         Metadata SQLite (tracking)
```

## Development Setup

### Prerequisites

- Nix (with flakes enabled)
- Docker and Docker Compose

### Quick Start

1. Enter the Nix development environment:
```bash
nix develop
```

2. Start Kreuzberg service:
```bash
docker-compose up -d
```

3. Run the server:
```bash
go run cmd/server/main.go
```

The server will start on `http://localhost:3000`

### Configuration

Copy `.env.example` to `.env` and adjust settings as needed:
```bash
cp .env.example .env
```

Configuration is loaded from environment variables. See `.env.example` for all available options.

## API Endpoints

### Health Check
```bash
curl http://localhost:3000/health
```

Response:
```json
{
  "status": "healthy",
  "kreuzberg_available": true,
  "gnucash_db_writable": true,
  "metadata_db_connected": true
}
```

### Upload Statement (Coming Soon)
```bash
curl -F "file=@statement.pdf" http://localhost:3000/upload
```

### List Statements (Coming Soon)
```bash
curl http://localhost:3000/statements
```

## Project Structure

```
moneymanager/
├── cmd/server/           # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── server/          # HTTP server and middleware
│   ├── statement/       # Statement processing
│   ├── kreuzberg/       # Kreuzberg API client
│   ├── transaction/     # Transaction normalization
│   ├── gnucash/         # GNU Cash library
│   └── database/        # Database access
├── data/                # Data directory (created at runtime)
├── logs/                # Log directory (created at runtime)
├── uploads/             # Temporary uploads (created at runtime)
└── tasks/               # Progress tracking markdown files
```

## Development

### Run with auto-reload
```bash
# Coming soon - currently run manually
go run cmd/server/main.go
```

### Run linters
```bash
golangci-lint run
```

### Format code
```bash
gofumpt -w .
```

### Run tests
```bash
go test ./...
```

## Implementation Status

See `tasks/` directory for detailed progress tracking:
- [Phase 1: Foundation](tasks/phase1-foundation.md) - In Progress
- Phase 2: Kreuzberg Integration - Pending
- Phase 3: GNU Cash Library (Read) - Pending
- Phase 4: GNU Cash Library (Write) - Pending
- Phase 5: Transaction Normalization - Pending
- Phase 6: Polish - Pending

## License

Private project
