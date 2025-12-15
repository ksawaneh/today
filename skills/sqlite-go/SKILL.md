# SQLite for Go Applications Skill

A comprehensive guide for using SQLite in Go applications, including pure-Go drivers, migrations, and best practices.

## When to Use This Skill

- Building local-first applications
- Single-file database storage
- CLI tools needing persistent storage
- Embedded databases without external dependencies
- Applications that need to work offline

---

## Driver Selection

### Pure Go: `modernc.org/sqlite`

**Pros:**
- No CGO required
- Easy cross-compilation
- Single binary deployment

**Cons:**
- ~10-20% slower than CGO version
- Larger binary size

```go
import (
    "database/sql"
    _ "modernc.org/sqlite"
)

db, err := sql.Open("sqlite", "path/to/database.db")
```

### CGO: `github.com/mattn/go-sqlite3`

**Pros:**
- Fastest performance
- Most feature-complete

**Cons:**
- Requires CGO (complex cross-compilation)
- Needs SQLite C library

```go
import (
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
)

db, err := sql.Open("sqlite3", "path/to/database.db")
```

### Recommendation

**Use `modernc.org/sqlite`** for CLI tools and TUI applications where cross-compilation simplicity matters more than raw performance.

---

## Database Setup

### XDG-Compliant Paths

```go
import "github.com/adrg/xdg"

func getDBPath() (string, error) {
    // Returns ~/.local/share/today/today.db on Linux
    // Returns ~/Library/Application Support/today/today.db on macOS
    // Returns %LOCALAPPDATA%\today\today.db on Windows
    dataDir, err := xdg.DataFile("today")
    if err != nil {
        return "", err
    }
    return filepath.Join(dataDir, "today.db"), nil
}
```

### Connection with Optimal Settings

```go
func OpenDB(path string) (*sql.DB, error) {
    db, err := sql.Open("sqlite", path)
    if err != nil {
        return nil, fmt.Errorf("opening database: %w", err)
    }
    
    // Enable WAL mode for better concurrency
    if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
        return nil, fmt.Errorf("enabling WAL: %w", err)
    }
    
    // Balance between safety and speed
    if _, err := db.Exec("PRAGMA synchronous=NORMAL"); err != nil {
        return nil, fmt.Errorf("setting synchronous: %w", err)
    }
    
    // Enable foreign key constraints
    if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
        return nil, fmt.Errorf("enabling foreign keys: %w", err)
    }
    
    // Wait up to 5 seconds if database is locked
    if _, err := db.Exec("PRAGMA busy_timeout=5000"); err != nil {
        return nil, fmt.Errorf("setting busy timeout: %w", err)
    }
    
    // Single connection for SQLite (avoid lock contention)
    db.SetMaxOpenConns(1)
    
    return db, nil
}
```

### PRAGMA Reference

| PRAGMA | Recommended | Purpose |
|--------|-------------|---------|
| `journal_mode=WAL` | Yes | Better concurrent read performance |
| `synchronous=NORMAL` | Yes | Balance safety/speed (use FULL for critical data) |
| `foreign_keys=ON` | Yes | Enforce referential integrity |
| `busy_timeout=5000` | Yes | Wait instead of immediate SQLITE_BUSY |
| `cache_size=-64000` | Optional | 64MB cache (negative = KB) |
| `temp_store=MEMORY` | Optional | Temp tables in RAM |

---

## Migrations

### Using Goose

```go
import (
    "embed"
    
    "github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func RunMigrations(db *sql.DB) error {
    goose.SetBaseFS(migrationsFS)
    goose.SetDialect("sqlite3")
    
    if err := goose.Up(db, "migrations"); err != nil {
        return fmt.Errorf("running migrations: %w", err)
    }
    
    return nil
}
```

### Migration File Structure

```
migrations/
├── 001_initial.sql
├── 002_add_habits.sql
└── 003_add_timer.sql
```

### Migration File Format

```sql
-- migrations/001_initial.sql
-- +goose Up
CREATE TABLE tasks (
    id TEXT PRIMARY KEY,
    text TEXT NOT NULL,
    project TEXT,
    done INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL,
    completed_at TEXT
);

CREATE INDEX idx_tasks_project ON tasks(project);
CREATE INDEX idx_tasks_done ON tasks(done);

-- +goose Down
DROP TABLE tasks;
```

```sql
-- migrations/002_add_habits.sql
-- +goose Up
CREATE TABLE habits (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    icon TEXT NOT NULL DEFAULT '✓',
    frequency TEXT NOT NULL DEFAULT 'daily',
    created_at TEXT NOT NULL
);

CREATE TABLE habit_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    habit_id TEXT NOT NULL REFERENCES habits(id) ON DELETE CASCADE,
    date TEXT NOT NULL,
    UNIQUE(habit_id, date)
);

CREATE INDEX idx_habit_logs_date ON habit_logs(date);

-- +goose Down
DROP TABLE habit_logs;
DROP TABLE habits;
```

---

## Repository Pattern

### Interface Definition

```go
// storage/storage.go
type TaskRepository interface {
    Create(ctx context.Context, task *Task) error
    GetByID(ctx context.Context, id string) (*Task, error)
    List(ctx context.Context, filter TaskFilter) ([]Task, error)
    Update(ctx context.Context, task *Task) error
    Delete(ctx context.Context, id string) error
}

type HabitRepository interface {
    Create(ctx context.Context, habit *Habit) error
    ToggleLog(ctx context.Context, habitID string, date time.Time) (bool, error)
    GetStreak(ctx context.Context, habitID string) (int, error)
    List(ctx context.Context) ([]Habit, error)
    Delete(ctx context.Context, id string) error
}
```

### SQLite Implementation

```go
// storage/sqlite/tasks.go
type SQLiteTaskRepo struct {
    db *sql.DB
}

func NewTaskRepo(db *sql.DB) *SQLiteTaskRepo {
    return &SQLiteTaskRepo{db: db}
}

func (r *SQLiteTaskRepo) Create(ctx context.Context, task *Task) error {
    query := `
        INSERT INTO tasks (id, text, project, done, created_at)
        VALUES (?, ?, ?, ?, ?)
    `
    
    _, err := r.db.ExecContext(ctx, query,
        task.ID,
        task.Text,
        task.Project,
        boolToInt(task.Done),
        task.CreatedAt.Format(time.RFC3339),
    )
    
    if err != nil {
        return fmt.Errorf("inserting task: %w", err)
    }
    
    return nil
}

func (r *SQLiteTaskRepo) GetByID(ctx context.Context, id string) (*Task, error) {
    query := `
        SELECT id, text, project, done, created_at, completed_at
        FROM tasks
        WHERE id = ?
    `
    
    var task Task
    var completedAt sql.NullString
    
    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &task.ID,
        &task.Text,
        &task.Project,
        &task.Done,
        &task.CreatedAt,
        &completedAt,
    )
    
    if err == sql.ErrNoRows {
        return nil, ErrNotFound
    }
    if err != nil {
        return nil, fmt.Errorf("querying task: %w", err)
    }
    
    if completedAt.Valid {
        t, _ := time.Parse(time.RFC3339, completedAt.String)
        task.CompletedAt = &t
    }
    
    return &task, nil
}

func (r *SQLiteTaskRepo) List(ctx context.Context, filter TaskFilter) ([]Task, error) {
    query := `
        SELECT id, text, project, done, created_at, completed_at
        FROM tasks
        WHERE 1=1
    `
    args := []interface{}{}
    
    if filter.Project != "" {
        query += " AND project = ?"
        args = append(args, filter.Project)
    }
    
    if filter.Done != nil {
        query += " AND done = ?"
        args = append(args, boolToInt(*filter.Done))
    }
    
    query += " ORDER BY created_at DESC"
    
    if filter.Limit > 0 {
        query += " LIMIT ?"
        args = append(args, filter.Limit)
    }
    
    rows, err := r.db.QueryContext(ctx, query, args...)
    if err != nil {
        return nil, fmt.Errorf("querying tasks: %w", err)
    }
    defer rows.Close()
    
    var tasks []Task
    for rows.Next() {
        var task Task
        var completedAt sql.NullString
        
        if err := rows.Scan(
            &task.ID,
            &task.Text,
            &task.Project,
            &task.Done,
            &task.CreatedAt,
            &completedAt,
        ); err != nil {
            return nil, fmt.Errorf("scanning task: %w", err)
        }
        
        if completedAt.Valid {
            t, _ := time.Parse(time.RFC3339, completedAt.String)
            task.CompletedAt = &t
        }
        
        tasks = append(tasks, task)
    }
    
    return tasks, rows.Err()
}
```

---

## Single Instance Locking

Prevent multiple instances from corrupting the database:

```go
import "github.com/gofrs/flock"

type App struct {
    db   *sql.DB
    lock *flock.Flock
}

func NewApp(dbPath string) (*App, error) {
    // Create lock file
    lockPath := dbPath + ".lock"
    fileLock := flock.New(lockPath)
    
    // Try to acquire exclusive lock
    locked, err := fileLock.TryLock()
    if err != nil {
        return nil, fmt.Errorf("acquiring lock: %w", err)
    }
    if !locked {
        return nil, fmt.Errorf("another instance is already running")
    }
    
    // Open database
    db, err := OpenDB(dbPath)
    if err != nil {
        fileLock.Unlock()
        return nil, err
    }
    
    return &App{db: db, lock: fileLock}, nil
}

func (a *App) Close() error {
    if err := a.db.Close(); err != nil {
        return err
    }
    return a.lock.Unlock()
}
```

---

## Backup and Recovery

### Automatic Backups

```go
type BackupManager struct {
    db        *sql.DB
    backupDir string
    maxBackups int
}

func NewBackupManager(db *sql.DB, backupDir string) *BackupManager {
    os.MkdirAll(backupDir, 0755)
    return &BackupManager{
        db:         db,
        backupDir:  backupDir,
        maxBackups: 7,
    }
}

func (bm *BackupManager) CreateBackup() error {
    timestamp := time.Now().Format("2006-01-02-150405")
    backupPath := filepath.Join(bm.backupDir, fmt.Sprintf("backup-%s.db", timestamp))
    
    // VACUUM INTO creates a consistent backup without locking
    _, err := bm.db.Exec(fmt.Sprintf("VACUUM INTO '%s'", backupPath))
    if err != nil {
        return fmt.Errorf("creating backup: %w", err)
    }
    
    // Rotate old backups
    return bm.rotateBackups()
}

func (bm *BackupManager) rotateBackups() error {
    entries, err := os.ReadDir(bm.backupDir)
    if err != nil {
        return err
    }
    
    var backups []string
    for _, entry := range entries {
        if strings.HasPrefix(entry.Name(), "backup-") && strings.HasSuffix(entry.Name(), ".db") {
            backups = append(backups, filepath.Join(bm.backupDir, entry.Name()))
        }
    }
    
    // Sort by name (timestamp makes them chronological)
    sort.Strings(backups)
    
    // Remove oldest backups if we exceed max
    for len(backups) > bm.maxBackups {
        if err := os.Remove(backups[0]); err != nil {
            return err
        }
        backups = backups[1:]
    }
    
    return nil
}

func (bm *BackupManager) RestoreLatest() error {
    entries, _ := os.ReadDir(bm.backupDir)
    
    var latest string
    for _, entry := range entries {
        if strings.HasPrefix(entry.Name(), "backup-") {
            path := filepath.Join(bm.backupDir, entry.Name())
            if path > latest {
                latest = path
            }
        }
    }
    
    if latest == "" {
        return fmt.Errorf("no backups found")
    }
    
    // Close current DB, copy backup, reopen
    // Implementation depends on your architecture
    return nil
}
```

### Scheduled Backups

```go
func (a *App) StartBackupScheduler(ctx context.Context) {
    ticker := time.NewTicker(24 * time.Hour)
    defer ticker.Stop()
    
    // Backup on startup
    a.backupManager.CreateBackup()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            if err := a.backupManager.CreateBackup(); err != nil {
                a.logger.Error("backup failed", "error", err)
            }
        }
    }
}
```

---

## Transactions

```go
func (r *SQLiteTaskRepo) CompleteMultiple(ctx context.Context, ids []string) error {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("beginning transaction: %w", err)
    }
    defer tx.Rollback()  // No-op if committed
    
    stmt, err := tx.PrepareContext(ctx, `
        UPDATE tasks SET done = 1, completed_at = ? WHERE id = ?
    `)
    if err != nil {
        return fmt.Errorf("preparing statement: %w", err)
    }
    defer stmt.Close()
    
    now := time.Now().Format(time.RFC3339)
    for _, id := range ids {
        if _, err := stmt.ExecContext(ctx, now, id); err != nil {
            return fmt.Errorf("updating task %s: %w", id, err)
        }
    }
    
    if err := tx.Commit(); err != nil {
        return fmt.Errorf("committing transaction: %w", err)
    }
    
    return nil
}
```

---

## Testing

### In-Memory Database

```go
func setupTestDB(t *testing.T) *sql.DB {
    t.Helper()
    
    db, err := sql.Open("sqlite", ":memory:")
    if err != nil {
        t.Fatalf("opening test db: %v", err)
    }
    
    // Run migrations
    if err := RunMigrations(db); err != nil {
        t.Fatalf("running migrations: %v", err)
    }
    
    t.Cleanup(func() {
        db.Close()
    })
    
    return db
}

func TestTaskRepo_Create(t *testing.T) {
    db := setupTestDB(t)
    repo := NewTaskRepo(db)
    
    task := &Task{
        ID:        "test-1",
        Text:      "Test task",
        CreatedAt: time.Now(),
    }
    
    err := repo.Create(context.Background(), task)
    if err != nil {
        t.Fatalf("creating task: %v", err)
    }
    
    got, err := repo.GetByID(context.Background(), "test-1")
    if err != nil {
        t.Fatalf("getting task: %v", err)
    }
    
    if got.Text != task.Text {
        t.Errorf("text = %q, want %q", got.Text, task.Text)
    }
}
```

---

## Quick Reference

```go
// Open with pure Go driver
import _ "modernc.org/sqlite"
db, _ := sql.Open("sqlite", "path.db")

// Essential PRAGMAs
db.Exec("PRAGMA journal_mode=WAL")
db.Exec("PRAGMA synchronous=NORMAL")
db.Exec("PRAGMA foreign_keys=ON")
db.Exec("PRAGMA busy_timeout=5000")

// Single connection for SQLite
db.SetMaxOpenConns(1)

// Backup
db.Exec("VACUUM INTO 'backup.db'")

// Check if table exists
db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", tableName)
```
