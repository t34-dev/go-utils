package pg

import (
	"context"
	"github.com/t34-dev/go-utils/pkg/db"

	"github.com/jackc/pgx/v4/pgxpool"
)

type ConnectFunc func(ctx context.Context) (*pgxpool.Pool, error)
type LogFunc func(ctx context.Context, q db.Query, args ...interface{})

type pgClient struct {
	masterDBC db.DB
}

func New(ctx context.Context, connector ConnectFunc, logger *LogFunc) (db.Client, error) {
	dbc, err := connector(ctx)
	if err != nil {
		return nil, err
	}

	masterDBC := NewDB(dbc, logger)
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
