package pg

import (
	"context"
	"errors"
	"github.com/t34-dev/go-utils/pkg/db"

	"github.com/jackc/pgx/v4/pgxpool"
)

type ConnectFunc func(ctx context.Context) (*pgxpool.Pool, error)
type LogFunc func(ctx context.Context, q db.Query, args ...interface{})

type pgClient struct {
	masterDBC db.DB
}

func New(pool *pgxpool.Pool, logger *LogFunc) (db.Client, error) {
	if pool == nil {
		return nil, errors.New("pool is nil")
	}

	masterDBC := NewDB(pool, logger)
	return &pgClient{
		masterDBC: masterDBC,
	}, nil
}

func (c *pgClient) DB() db.DB {
	return c.masterDBC
}

func (c *pgClient) Close() error {
	if c.masterDBC != nil {
		c.masterDBC.Close()
	}

	return nil
}
