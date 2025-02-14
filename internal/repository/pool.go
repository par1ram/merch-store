// repository/pool.go
package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// PoolIface — интерфейс для пула соединений.
// Обратите внимание, что здесь Begin возвращает pgx.Tx.
type PoolIface interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	// Если нужны и другие методы, добавьте их здесь.
}
