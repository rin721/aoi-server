package ports

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrNotFound                  = errors.New("record not found")
	ErrNilTxFunc                 = errors.New("transaction function is nil")
	ErrInvalidTxOptions          = errors.New("invalid transaction options")
	ErrNestedTransactionDisabled = errors.New("nested transaction is disabled")
)

type Database interface {
	Executor
	Close() error
	Ping(context.Context) error
	SQLDB() (*sql.DB, error)
	WithTx(context.Context, TxFunc) error
	WithTxOptions(context.Context, *TxOptions, TxFunc) error
}

type Executor interface {
	Create(context.Context, any) error
	Save(context.Context, any) error
	First(context.Context, any, ...QueryOption) error
	Find(context.Context, any, ...QueryOption) error
	Update(context.Context, any, map[string]any, ...QueryOption) (Result, error)
	Delete(context.Context, any, ...QueryOption) (Result, error)
	Exec(context.Context, string, ...any) (Result, error)
	Raw(context.Context, any, string, ...any) (Result, error)
	Count(context.Context, any, ...QueryOption) (int64, error)
	HasTable(context.Context, any) (bool, error)
}

type Result struct {
	RowsAffected int64
}

type Query struct {
	Table       string
	Where       []Condition
	Order       string
	Limit       int
	Offset      int
	Unscoped    bool
	WithDeleted bool
}

type Condition struct {
	Expr string
	Args []any
}

type QueryOption func(*Query)

func Table(name string) QueryOption {
	return func(q *Query) {
		q.Table = name
	}
}

func Where(expr string, args ...any) QueryOption {
	return func(q *Query) {
		q.Where = append(q.Where, Condition{Expr: expr, Args: args})
	}
}

func Order(expr string) QueryOption {
	return func(q *Query) {
		q.Order = expr
	}
}

func Limit(n int) QueryOption {
	return func(q *Query) {
		q.Limit = n
	}
}

func Offset(n int) QueryOption {
	return func(q *Query) {
		q.Offset = n
	}
}

type TxFunc func(context.Context, Executor) error

type TxOptions struct {
	Isolation                sql.IsolationLevel
	ReadOnly                 bool
	Timeout                  time.Duration
	DisableNestedTransaction bool
}
