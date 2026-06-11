// Package migrator wraps goose migrations behind a small project-owned API.
package migrator

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"sync"

	"github.com/pressly/goose/v3"
	"github.com/rei0721/go-scaffold/pkg/database"
)

const DefaultDir = "internal/migrations"

type SQLProvider interface {
	SQLDB() (*sql.DB, error)
}

type Config struct {
	Driver string
	Dir    string
}

type Runner interface {
	Up(context.Context) error
	Down(context.Context) error
	Status(context.Context, io.Writer) error
}

type runner struct {
	db  *sql.DB
	cfg Config
}

var gooseMu sync.Mutex

func New(provider SQLProvider, cfg Config) (Runner, error) {
	if provider == nil {
		return nil, fmt.Errorf("migrator database provider is nil")
	}
	db, err := provider.SQLDB()
	if err != nil {
		return nil, err
	}
	if cfg.Dir == "" {
		cfg.Dir = DefaultDir
	}
	if cfg.Driver == "" {
		return nil, fmt.Errorf("migrator driver is required")
	}
	return &runner{db: db, cfg: cfg}, nil
}

func (r *runner) Up(ctx context.Context) error {
	return r.run(ctx, nil, func(ctx context.Context) error {
		return goose.UpContext(ctx, r.db, r.cfg.Dir, goose.WithNoColor(true))
	})
}

func (r *runner) Down(ctx context.Context) error {
	return r.run(ctx, nil, func(ctx context.Context) error {
		return goose.DownContext(ctx, r.db, r.cfg.Dir, goose.WithNoColor(true))
	})
}

func (r *runner) Status(ctx context.Context, w io.Writer) error {
	return r.run(ctx, w, func(ctx context.Context) error {
		return goose.StatusContext(ctx, r.db, r.cfg.Dir, goose.WithNoColor(true))
	})
}

func (r *runner) run(ctx context.Context, w io.Writer, fn func(context.Context) error) error {
	if ctx == nil {
		ctx = context.Background()
	}
	gooseMu.Lock()
	defer gooseMu.Unlock()
	if err := goose.SetDialect(gooseDialect(r.cfg.Driver)); err != nil {
		return err
	}
	if w != nil {
		goose.SetLogger(writerLogger{w: w})
	} else {
		goose.SetLogger(writerLogger{w: io.Discard})
	}
	return fn(ctx)
}

func gooseDialect(driver string) string {
	switch driver {
	case string(database.DriverSQLite), "sqlite3":
		return "sqlite3"
	case string(database.DriverPostgres), "postgresql":
		return "postgres"
	default:
		return driver
	}
}

type writerLogger struct {
	w io.Writer
}

func (l writerLogger) Fatalf(format string, v ...interface{}) {
	_, _ = fmt.Fprintf(l.w, format+"\n", v...)
}

func (l writerLogger) Printf(format string, v ...interface{}) {
	_, _ = fmt.Fprintf(l.w, format+"\n", v...)
}
