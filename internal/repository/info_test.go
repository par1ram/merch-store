package repository_test

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"

	"github.com/par1ram/merch-store/internal/db"
	"github.com/par1ram/merch-store/internal/repository"
	"github.com/par1ram/merch-store/internal/utils"
)

func TestInfoRepository_GetCoins_Success(t *testing.T) {
	// Создаем моковый пул.
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	// Создаем экземпляр db.Queries с использованием мокового пула.
	queries := db.New(mockPool)

	// Для запроса GetEmployeeByID ожидается выборка столбцов:
	// id, username, coins, created_at
	// При этом coins – int32.
	// Создадим dummy created_at (например, текущее время).
	now := time.Now()
	createdAt := pgtype.Timestamptz{Time: now, Valid: true}

	rows := pgxmock.NewRows([]string{"id", "username", "coins", "created_at"}).
		AddRow(int32(123), "testuser", int32(150), createdAt)

	// Ожидаем запрос, сгенерированный sqlc (точное совпадение может зависеть от форматирования SQL).
	mockPool.ExpectQuery(regexp.QuoteMeta(`
SELECT
  id,
  username,
  coins,
  created_at
FROM employees
WHERE id = $1
`)).
		WithArgs(int32(123)).
		WillReturnRows(rows)

	repoInstance := repository.NewInfoRepository(queries, utils.NewLogger())

	coins, err := repoInstance.GetCoins(context.Background(), 123)
	assert.NoError(t, err)
	assert.Equal(t, 150, coins)

	err = mockPool.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestInfoRepository_GetInventory_Success(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	queries := db.New(mockPool)

	// Для запроса GetInventoryByEmployeeID ожидаем столбцы:
	// employee_id, merch_id, merch_name, quantity
	rows := pgxmock.NewRows([]string{"employee_id", "merch_id", "merch_name", "quantity"}).
		AddRow(int32(123), int32(1), "T-Shirt", int32(2)).
		AddRow(int32(123), int32(2), "Hoodie", int32(1))

	mockPool.ExpectQuery(regexp.QuoteMeta(`
SELECT
  i.employee_id,
  i.merch_id,
  m.name AS merch_name,
  i.quantity
FROM inventory i
JOIN merch m ON i.merch_id = m.id
WHERE i.employee_id = $1
ORDER BY m.name
`)).
		WithArgs(int32(123)).
		WillReturnRows(rows)

	repoInstance := repository.NewInfoRepository(queries, utils.NewLogger())
	inventory, err := repoInstance.GetInventory(context.Background(), 123)
	assert.NoError(t, err)

	expectedInventory := []repository.InventoryItem{
		{Type: "T-Shirt", Quantity: 2},
		{Type: "Hoodie", Quantity: 1},
	}
	assert.Equal(t, expectedInventory, inventory)

	err = mockPool.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestInfoRepository_GetReceivedTransfers_Success(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	queries := db.New(mockPool)

	// Для запроса GetReceivedTransfers ожидаем столбцы:
	// amount, from_user
	rows := pgxmock.NewRows([]string{"amount", "from_user"}).
		AddRow(int32(50), "Alice").
		AddRow(int32(30), "Charlie")

	// В запросе используется аргумент типа pgtype.Int4.
	arg := pgtype.Int4{Int32: int32(123), Valid: true}
	mockPool.ExpectQuery(regexp.QuoteMeta(`
SELECT
  ct.amount,
  e.username AS from_user
FROM coin_transactions ct
JOIN employees e ON ct.from_employee_id = e.id
WHERE ct.transaction_type = 'transfer'
  AND ct.to_employee_id = $1
ORDER BY ct.created_at DESC
`)).
		WithArgs(arg).
		WillReturnRows(rows)

	repoInstance := repository.NewInfoRepository(queries, utils.NewLogger())
	received, err := repoInstance.GetReceivedTransfers(context.Background(), 123)
	assert.NoError(t, err)

	expectedReceived := []repository.ReceivedTransaction{
		{FromUser: "Alice", Amount: 50},
		{FromUser: "Charlie", Amount: 30},
	}
	assert.Equal(t, expectedReceived, received)

	err = mockPool.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestInfoRepository_GetSentTransfers_Success(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	queries := db.New(mockPool)

	// Для запроса GetSentTransfers ожидаем столбцы:
	// amount, to_user
	rows := pgxmock.NewRows([]string{"amount", "to_user"}).
		AddRow(int32(40), "Bob").
		AddRow(int32(20), "David")

	mockPool.ExpectQuery(regexp.QuoteMeta(`
SELECT
  ct.amount,
  e.username AS to_user
FROM coin_transactions ct
JOIN employees e ON ct.to_employee_id = e.id
WHERE ct.transaction_type = 'transfer'
  AND ct.from_employee_id = $1
ORDER BY ct.created_at DESC
`)).
		WithArgs(int32(123)).
		WillReturnRows(rows)

	repoInstance := repository.NewInfoRepository(queries, utils.NewLogger())
	sent, err := repoInstance.GetSentTransfers(context.Background(), 123)
	assert.NoError(t, err)

	expectedSent := []repository.SentTransaction{
		{ToUser: "Bob", Amount: 40},
		{ToUser: "David", Amount: 20},
	}
	assert.Equal(t, expectedSent, sent)

	err = mockPool.ExpectationsWereMet()
	assert.NoError(t, err)
}
