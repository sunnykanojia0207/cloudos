// Package database provides the built-in database.sqlite provider backed by
// SQLite. It implements the Database capability interface.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/cloudos/cloudos/capabilities"
	"github.com/cloudos/cloudos/providers"

	_ "github.com/mattn/go-sqlite3"
)

const (
	providerName    = "database.sqlite"
	providerVersion = "0.1.0"
	providerDesc    = "Built-in SQLite database provider"
)

// SQLiteProvider implements the Database capability backed by SQLite.
type SQLiteProvider struct {
	mu      sync.Mutex
	state   providers.State
	db      *sql.DB
	dsn     string
}

// NewSQLiteProvider creates a new SQLite database provider.
func NewSQLiteProvider(dsn string) *SQLiteProvider {
	return &SQLiteProvider{
		state: providers.StateDiscovered,
		dsn:   dsn,
	}
}

// Info returns provider metadata.
func (p *SQLiteProvider) Info() providers.Info {
	return providers.Info{
		Name:        providerName,
		Version:     providerVersion,
		Description: providerDesc,
		Capability:  "database",
	}
}

// Init configures the provider.
func (p *SQLiteProvider) Init(ctx context.Context, config map[string]interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.state = providers.StateInit
	return nil
}

// Start opens the SQLite database connection.
func (p *SQLiteProvider) Start(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	db, err := sql.Open("sqlite3", p.dsn)
	if err != nil {
		p.state = providers.StateFailed
		return fmt.Errorf("open sqlite: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		p.state = providers.StateFailed
		return fmt.Errorf("ping sqlite: %w", err)
	}

	// Enable WAL mode for better concurrent performance.
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return fmt.Errorf("enable WAL: %w", err)
	}

	p.db = db
	p.state = providers.StateReady
	return nil
}

// Stop closes the database connection.
func (p *SQLiteProvider) Stop(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.db != nil {
		if err := p.db.Close(); err != nil {
			return fmt.Errorf("close sqlite: %w", err)
		}
		p.db = nil
	}
	p.state = providers.StateStopped
	return nil
}

// State returns the current provider state.
func (p *SQLiteProvider) State() providers.State {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.state
}

// Capability returns the database capability.
func (p *SQLiteProvider) Capability() capabilities.Capability {
	return p
}

// --- Capability interface implementation -----------------------------------

func (p *SQLiteProvider) ID() capabilities.ID          { return "database" }
func (p *SQLiteProvider) Version() capabilities.Version { return capabilities.Version{Major: 1, Minor: 0, Patch: 0} }

func (p *SQLiteProvider) Validate(ctx context.Context) error {
	if p.state != providers.StateReady {
		return fmt.Errorf("provider not ready")
	}
	return p.db.PingContext(ctx)
}

func (p *SQLiteProvider) Exec(ctx context.Context, query string, args ...interface{}) (capabilities.Result, error) {
	res, err := p.db.ExecContext(ctx, query, args...)
	if err != nil {
		return capabilities.Result{}, fmt.Errorf("exec: %w", err)
	}

	rows, _ := res.RowsAffected()
	id, _ := res.LastInsertId()

	return capabilities.Result{
		RowsAffected: rows,
		LastInsertID: id,
	}, nil
}

func (p *SQLiteProvider) Query(ctx context.Context, query string, args ...interface{}) (capabilities.Rows, error) {
	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return capabilities.Rows{}, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return capabilities.Rows{}, fmt.Errorf("columns: %w", err)
	}

	var result capabilities.Rows
	result.Columns = columns

	for rows.Next() {
		values := make([]interface{}, len(columns))
		ptrs := make([]interface{}, len(columns))
		for i := range values {
			ptrs[i] = &values[i]
		}

		if err := rows.Scan(ptrs...); err != nil {
			return capabilities.Rows{}, fmt.Errorf("scan: %w", err)
		}

		row := make(map[string]interface{}, len(columns))
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				val = string(b)
			}
			row[col] = val
		}
		result.Rows = append(result.Rows, row)
	}

	if result.Rows == nil {
		result.Rows = []map[string]interface{}{}
	}

	return result, rows.Err()
}

func (p *SQLiteProvider) Migrate(ctx context.Context, migrations []capabilities.Migration) error {
	for _, m := range migrations {
		if _, err := p.db.ExecContext(ctx, m.Query); err != nil {
			return fmt.Errorf("migration %q: %w", m.ID, err)
		}
	}
	return nil
}

func (p *SQLiteProvider) Ping(ctx context.Context) error {
	return p.db.PingContext(ctx)
}
