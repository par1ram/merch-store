package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type PoolIface interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}
