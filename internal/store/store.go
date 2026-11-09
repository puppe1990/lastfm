package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/puppe1990/amarra-cais/pkg/cais/devlog"
	"github.com/puppe1990/amarra-cais/pkg/cais/session"
	caissqlite "github.com/puppe1990/amarra-cais/pkg/cais/sqlite"
	"github.com/puppe1990/amarra-cais/pkg/cais/sqllog"
	_ "modernc.org/sqlite"
)

type Store interface {
	Sessions() session.Store
	Ping() error
	DB() *sql.DB
	Close() error
}

type SQLiteStore struct {
	db *sqllog.DB
}

func NewSQLiteStore(dsn string, env string) (*SQLiteStore, error) {
	if dsn != ":memory:" {
		dir := filepath.Dir(dsn)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create db dir: %w", err)
		}
	}

	db, err := sql.Open("sqlite", caissqlite.DSN(dsn))
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := applyMigrations(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := session.EnsureSQLiteSchema(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("sessions schema: %w", err)
	}

	cfg := sqllog.ConfigForEnv(env)
	if cfg.Enabled {
		cfg.Writer = devlog.MirrorDefault(os.Stdout)
	}
	return &SQLiteStore{db: sqllog.Wrap(db, cfg)}, nil
}

func (s *SQLiteStore) Sessions() session.Store {
	return session.NewSQLiteStore(s.db.Raw())
}

func (s *SQLiteStore) Ping() error {
	return s.db.Raw().Ping()
}

func (s *SQLiteStore) DB() *sql.DB {
	return s.db.Raw()
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}
