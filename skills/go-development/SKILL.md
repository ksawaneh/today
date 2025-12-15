# Go Development Skill

A comprehensive guide for building production-quality Go applications.

## When to Use This Skill

- Building any Go application (CLI, TUI, API, library)
- Setting up new Go projects
- Debugging Go code
- Writing tests
- Cross-compilation and distribution

---

## Project Structure

### Standard Layout

```
myapp/
├── cmd/
│   └── myapp/
│       └── main.go          # Entrypoint (minimal, calls internal/)
├── internal/                # Private application code
│   ├── app/                 # Application lifecycle
│   ├── config/              # Configuration loading
│   ├── storage/             # Database/file operations
│   └── service/             # Business logic
├── pkg/                     # Public libraries (if any)
├── migrations/              # SQL migrations (embed with //go:embed)
├── testdata/                # Test fixtures
├── go.mod
├── go.sum
├── Makefile                 # Build automation
└── README.md
```

### Key Principles

1. **`cmd/` is minimal** — Only parse flags, load config, call `internal/`
2. **`internal/` is private** — Cannot be imported by other modules
3. **`pkg/` is public** — Only if you intend others to import it
4. **Flat is better than nested** — Don't over-organize prematurely

---

## Essential Patterns

### Error Handling

```go
// ALWAYS wrap errors with context
if err != nil {
    return fmt.Errorf("loading config from %s: %w", path, err)
}

// Check specific errors with errors.Is / errors.As
if errors.Is(err, os.ErrNotExist) {
    // Handle missing file
}

// Define sentinel errors for expected conditions
var ErrNotFound = errors.New("not found")

// Custom error types for rich errors
type ValidationError struct {
    Field   string
    Message string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("%s: %s", e.Field, e.Message)
}
```

### Configuration

```go
// Use struct tags for environment variables
type Config struct {
    Port     int    `env:"PORT" envDefault:"8080"`
    Database string `env:"DATABASE_URL,required"`
    Debug    bool   `env:"DEBUG" envDefault:"false"`
}

// Load with github.com/caarlos0/env/v10
func LoadConfig() (*Config, error) {
    cfg := &Config{}
    if err := env.Parse(cfg); err != nil {
        return nil, fmt.Errorf("parsing config: %w", err)
    }
    return cfg, nil
}
```

### Dependency Injection

```go
// Define interfaces where they're USED, not where they're implemented
type Storage interface {
    SaveTask(ctx context.Context, task Task) error
    GetTask(ctx context.Context, id string) (Task, error)
}

// Constructors take dependencies
func NewService(storage Storage, logger *slog.Logger) *Service {
    return &Service{
        storage: storage,
        logger:  logger,
    }
}
```

### Context Usage

```go
// Always pass context as first parameter
func (s *Service) DoWork(ctx context.Context, input Input) error {
    // Check for cancellation
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }
    
    // Pass context to downstream calls
    return s.storage.Save(ctx, input)
}
```

---

## Testing

### Table-Driven Tests

```go
func TestParseTask(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    Task
        wantErr bool
    }{
        {
            name:  "simple task",
            input: "Buy groceries",
            want:  Task{Text: "Buy groceries"},
        },
        {
            name:  "task with priority",
            input: "!high Fix bug",
            want:  Task{Text: "Fix bug", Priority: High},
        },
        {
            name:    "empty input",
            input:   "",
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParseTask(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("ParseTask() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("ParseTask() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Test Helpers

```go
// t.Helper() marks function as helper (better error locations)
func setupTestDB(t *testing.T) *sql.DB {
    t.Helper()
    
    db, err := sql.Open("sqlite", ":memory:")
    if err != nil {
        t.Fatalf("opening test db: %v", err)
    }
    
    t.Cleanup(func() {
        db.Close()
    })
    
    return db
}
```

### Golden File Tests

```go
func TestRender(t *testing.T) {
    got := Render(input)
    
    golden := filepath.Join("testdata", t.Name()+".golden")
    
    if *update {
        os.WriteFile(golden, []byte(got), 0644)
    }
    
    want, _ := os.ReadFile(golden)
    if got != string(want) {
        t.Errorf("output mismatch:\n%s", diff(want, got))
    }
}
```

---

## Common Pitfalls

### 1. Nil Map Write

```go
// PANIC: assignment to entry in nil map
var m map[string]int
m["key"] = 1

// FIX: Initialize the map
m := make(map[string]int)
m["key"] = 1
```

### 2. Goroutine Leak

```go
// LEAK: goroutine blocked forever if nobody receives
func bad() chan int {
    ch := make(chan int)
    go func() {
        ch <- expensiveOperation()  // Blocks if ignored
    }()
    return ch
}

// FIX: Use buffered channel or context cancellation
func good() chan int {
    ch := make(chan int, 1)  // Buffered
    go func() {
        ch <- expensiveOperation()
    }()
    return ch
}
```

### 3. Defer in Loop

```go
// BAD: All defers run at function end, not loop iteration end
for _, file := range files {
    f, _ := os.Open(file)
    defer f.Close()  // Won't close until function returns!
}

// FIX: Use closure or explicit close
for _, file := range files {
    func() {
        f, _ := os.Open(file)
        defer f.Close()
        // Process file
    }()
}
```

### 4. Range Variable Capture

```go
// BUG (pre-Go 1.22): All goroutines see same 'item'
for _, item := range items {
    go func() {
        process(item)  // Always processes last item!
    }()
}

// FIX (pre-Go 1.22): Shadow the variable
for _, item := range items {
    item := item  // Shadow
    go func() {
        process(item)
    }()
}

// Go 1.22+: Fixed by default
```

---

## Build and Distribution

### Makefile

```makefile
BINARY := myapp
VERSION := $(shell git describe --tags --always --dirty)
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: build test lint clean

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/$(BINARY)

test:
	go test -race -cover ./...

lint:
	golangci-lint run

clean:
	rm -f $(BINARY)

install:
	go install -ldflags "$(LDFLAGS)" ./cmd/$(BINARY)
```

### Cross-Compilation

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o myapp-linux-amd64 ./cmd/myapp

# macOS ARM64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o myapp-darwin-arm64 ./cmd/myapp

# Windows
GOOS=windows GOARCH=amd64 go build -o myapp-windows-amd64.exe ./cmd/myapp
```

### Embedding Files

```go
import "embed"

//go:embed migrations/*.sql
var migrationsFS embed.FS

//go:embed templates/*
var templatesFS embed.FS

// Access embedded files
data, err := migrationsFS.ReadFile("migrations/001_init.sql")
```

---

## Useful Libraries

| Category | Library | Purpose |
|----------|---------|---------|
| CLI | `github.com/spf13/cobra` | Command-line apps |
| Config | `github.com/caarlos0/env/v10` | Environment variables |
| Logging | `log/slog` (stdlib) | Structured logging |
| Testing | `github.com/stretchr/testify` | Assertions, mocks |
| SQL | `github.com/jmoiron/sqlx` | SQL extensions |
| HTTP | `github.com/go-chi/chi/v5` | Lightweight router |
| Validation | `github.com/go-playground/validator/v10` | Struct validation |

---

## Quick Reference

```bash
# Initialize module
go mod init github.com/user/project

# Add dependency
go get github.com/user/package@latest

# Tidy dependencies
go mod tidy

# Run tests
go test ./...

# Run with race detector
go test -race ./...

# Build with version
go build -ldflags "-X main.version=1.0.0" ./cmd/myapp

# View documentation
go doc fmt.Sprintf

# Format code
gofmt -w .

# Vet code
go vet ./...
```
