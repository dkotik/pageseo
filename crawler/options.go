package crawler

import (
	"errors"
	"log/slog"
	"time"

	"github.com/dkotik/pageseo/crawler/repository"
	"zombiezen.com/go/sqlite"
)

type options struct {
	// Filter  Filter
	SQLiteConn *sqlite.Conn
	Repository repository.Repository
	TimeToLive time.Duration
	BatchSize  int
	Logger     *slog.Logger
}

type Option func(options) (options, error)

func WithSQLiteConn(conn *sqlite.Conn) Option {
	return func(o options) (options, error) {
		if conn == nil {
			return o, errors.New("nil connection")
		}
		if o.SQLiteConn != nil {
			return o, errors.New("connection is already set")
		}
		o.SQLiteConn = conn
		return o, nil
	}
}

func WithRepository(repo repository.Repository) Option {
	return func(o options) (options, error) {
		if repo == nil {
			return o, errors.New("nil repository")
		}
		if o.Repository != nil {
			return o, errors.New("repository is already set")
		}
		if o.SQLiteConn != nil {
			return o, errors.New("SQLite connection conflicts with the repository option")
		}
		o.Repository = repo
		return o, nil
	}
}

func WithTimeToLive(d time.Duration) Option {
	return func(o options) (options, error) {
		if d == 0 {
			return o, errors.New("invalid time to live")
		}
		if o.TimeToLive != 0 {
			return o, errors.New("time to live is already set")
		}
		o.TimeToLive = d
		return o, nil
	}
}

func WithBatchSize(size int) Option {
	return func(o options) (options, error) {
		if size == 0 {
			return o, errors.New("invalid batch size")
		}
		if o.BatchSize != 0 {
			return o, errors.New("batch size is already set")
		}
		o.BatchSize = size
		return o, nil
	}
}

func WithLogger(logger *slog.Logger) Option {
	return func(o options) (options, error) {
		if logger == nil {
			return o, errors.New("nil logger")
		}
		if o.Logger != nil {
			return o, errors.New("logger is already set")
		}
		o.Logger = logger
		return o, nil
	}
}
