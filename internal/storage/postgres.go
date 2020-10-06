package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/lib/pq"

	"github.com/pavelmemory/faceit-users/internal"
)

type ExecResult interface {
	Err() error
	Affected() int64
}

type MultiResult interface {
	Next() bool
	Scan(dst ...interface{}) error
	Close() error
}

type SingleResult interface {
	Scan(dst ...interface{}) error
}

type Runner interface {
	Exec(ctx context.Context, query string, params ...interface{}) ExecResult
	Query(ctx context.Context, query string, params ...interface{}) (MultiResult, error)
	QuerySingle(ctx context.Context, query string, params ...interface{}) SingleResult
}

// NewPostgres returns a connection pool ready to execute statements on PostgreSQL database.
// TODO: there should be a PgBouncer instance between clients and PostgreSQL dabatase.
func NewPostgres(addr string, pwd string) (*Postgres, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("split host and port: %w", err)
	}

	// TODO: make more dynamic configuration of the database connection
	db, err := sql.Open("postgres", "host="+host+" port="+port+" user=postgres password="+pwd+" dbname=postgres sslmode=disable binary_parameters=yes")
	if err != nil {
		return nil, fmt.Errorf("open connection: %w", err)
	}

	// TODO: should be weighted if those are meaningful
	db.SetMaxOpenConns(16)
	db.SetMaxIdleConns(4)
	db.SetConnMaxLifetime(30 * time.Second)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping databse: %w", err)
	}

	return &Postgres{db: db}, nil
}

type Postgres struct {
	db *sql.DB
}

func (p *Postgres) WithTx(ctx context.Context, action func(runner Runner) error) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := action(txRunner{tx: tx}); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (p *Postgres) WithoutTx(_ context.Context, action func(runner Runner) error) error {
	return action(qRunner{db: p.db})
}

func (p *Postgres) Close() {
	p.db.Close()
}

type qRunner struct {
	db *sql.DB
}

func (q qRunner) Exec(ctx context.Context, query string, params ...interface{}) ExecResult {
	res, err := q.db.ExecContext(ctx, query, params...)
	if err != nil {
		return execResult{err: err}
	}

	n, err := res.RowsAffected()
	return execResult{result: n, err: err}
}

func (q qRunner) Query(ctx context.Context, query string, params ...interface{}) (MultiResult, error) {
	return q.db.QueryContext(ctx, query, params...)
}

func (q qRunner) QuerySingle(ctx context.Context, query string, params ...interface{}) SingleResult {
	return q.db.QueryRowContext(ctx, query, params...)
}

type txRunner struct {
	tx *sql.Tx
}

func (r txRunner) Exec(ctx context.Context, query string, params ...interface{}) ExecResult {
	res, err := r.tx.ExecContext(ctx, query, params...)
	if err != nil {
		return execResult{err: err}
	}

	n, err := res.RowsAffected()
	return execResult{result: n, err: err}
}

func (r txRunner) Query(ctx context.Context, query string, params ...interface{}) (MultiResult, error) {
	return r.tx.QueryContext(ctx, query, params...)
}

func (r txRunner) QuerySingle(ctx context.Context, query string, params ...interface{}) SingleResult {
	return r.tx.QueryRowContext(ctx, query, params...)
}

type execResult struct {
	result int64
	err    error
}

func (r execResult) Err() error {
	return r.err
}

func (r execResult) Affected() int64 {
	return r.result
}

func convertError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return internal.ErrNotFound
	}

	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		var cause error
		// https://www.postgresql.org/docs/11/errcodes-appendix.html
		switch pqErr.Code {
		case "23000", "23001", "23502", "23503", "23514", "23P01":
			cause = internal.ErrBadInput
		case "23505":
			cause = internal.ErrNotUnique
		default:
			return err
		}

		return fmt.Errorf("%v: %w", err, cause)
	}

	return err
}
