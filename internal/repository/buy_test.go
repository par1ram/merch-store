package repository_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"

	"github.com/par1ram/merch-store/internal/db"
	"github.com/par1ram/merch-store/internal/repository"
	"github.com/par1ram/merch-store/internal/utils"
)

func TestBuyRepository_GetMerch_Success(t *testing.T) {
	// Создаем мок для пула.
	mockPool, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create pgxmock pool: %v", err)
	}
	defer mockPool.Close()

	// Создаем экземпляр db.Queries, используя моковый пул.
	queries := db.New(mockPool)

	// Настраиваем ожидаемые строки, которые вернет запрос GetMerchByName.
	rows := pgxmock.NewRows([]string{"id", "name", "price"}).
		AddRow(int32(1), "T-Shirt", int32(100))

	// Настраиваем ожидание запроса.
	mockPool.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, price FROM merch WHERE name = $1`)).
		WithArgs("T-Shirt").
		WillReturnRows(rows)

	// Создаем репозиторий через публичный конструктор.
	repo := repository.NewBuyRepository(mockPool, queries, utils.NewLogger())

	// Вызываем тестируемый метод.
	merch, err := repo.GetMerch(context.Background(), "T-Shirt")
	assert.NoError(t, err)
	assert.Equal(t, int32(1), merch.ID)
	assert.Equal(t, "T-Shirt", merch.Name)
	assert.Equal(t, int32(100), merch.Price)

	// Проверяем, что все ожидания мокового пула выполнены.
	err = mockPool.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestBuyRepository_GetMerch_NotFound(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	queries := db.New(mockPool)
	logger := utils.NewLogger()
	repo := repository.NewBuyRepository(mockPool, queries, logger)

	ctx := context.Background()
	item := "NonExistent"

	// Настраиваем пустую выборку => sql.ErrNoRows.
	mockPool.
		ExpectQuery(regexp.QuoteMeta(`SELECT id, name, price FROM merch WHERE name = $1`)).
		WithArgs(item).
		WillReturnRows(mockPool.NewRows([]string{"id", "name", "price"})) // без строк

	merch, err := repo.GetMerch(ctx, item)
	assert.Error(t, err)
	assert.Equal(t, int32(0), merch.ID)
	assert.Empty(t, merch.Name)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

// 3. GetMerch: ошибка запроса.
func TestBuyRepository_GetMerch_Error(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	queries := db.New(mockPool)
	logger := utils.NewLogger()
	repo := repository.NewBuyRepository(mockPool, queries, logger)

	ctx := context.Background()
	item := "AnyItem"

	mockPool.
		ExpectQuery(regexp.QuoteMeta(`SELECT id, name, price FROM merch WHERE name = $1`)).
		WithArgs(item).
		WillReturnError(assert.AnError) // имитация ошибки

	merch, err := repo.GetMerch(ctx, item)
	assert.Error(t, err)
	assert.Empty(t, merch)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

// 4. GetBalance: успешное получение баланса.
func TestBuyRepository_GetBalance_Success(t *testing.T) {
	// Создаём моковый пул.
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	queries := db.New(mockPool)
	logger := utils.NewLogger()
	repo := repository.NewBuyRepository(mockPool, queries, logger)

	ctx := context.Background()
	userID := int32(123)

	// Возвращаем одну строку с coins=500
	rows := mockPool.NewRows([]string{"coins"}).AddRow(int32(500))

	// Используем "гибкую" регулярку с (?s).* и без пробелов, чтобы
	// совпадало с sqlc-шным запросом: "-- name: GetCoinsByID :one\nSELECT coins FROM employees WHERE id=$1"
	mockPool.ExpectQuery(`(?s).*SELECT coins FROM employees WHERE id=\$1.*`).
		WithArgs(userID).
		WillReturnRows(rows)

	balance, err := repo.GetBalance(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, int32(500), balance)

	assert.NoError(t, mockPool.ExpectationsWereMet())
}

// 5. GetBalance: ошибка при запросе.
func TestBuyRepository_GetBalance_Error(t *testing.T) {
	// Создаём моковый пул.
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	queries := db.New(mockPool)
	logger := utils.NewLogger()
	repo := repository.NewBuyRepository(mockPool, queries, logger)

	ctx := context.Background()
	userID := int32(999)

	// Имитация ошибки, например, "no rows found" или любая другая:
	mockPool.ExpectQuery(`(?s).*SELECT coins FROM employees WHERE id=\$1.*`).
		WithArgs(userID).
		WillReturnError(assert.AnError)

	balance, err := repo.GetBalance(ctx, userID)
	assert.Error(t, err)
	assert.Equal(t, int32(0), balance)

	assert.NoError(t, mockPool.ExpectationsWereMet())
}

// 6. DeductCoins: успешное списание.
// По коду repo.DeductCoins просто вызывает UpdateEmployeeCoins(...),
// и при успехе возвращает (1, nil). При ошибке — (0, err).
func TestBuyRepository_DeductCoins_Success(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	queries := db.New(mockPool)
	logger := utils.NewLogger()
	repo := repository.NewBuyRepository(mockPool, queries, logger)

	ctx := context.Background()
	userID := int32(100)
	amount := int32(50)

	// Ожидаем exec UpdateEmployeeCoins, где coins = -amount
	mockPool.
		ExpectExec(regexp.QuoteMeta(`UPDATE employees SET coins = coins + $2 WHERE id = $1 AND coins + $2 >= 0`)).
		WithArgs(userID, -amount).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	affected, err := repo.DeductCoins(ctx, userID, amount)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), affected)

	assert.NoError(t, mockPool.ExpectationsWereMet())
}

// 7. DeductCoins: ошибка при списании.
func TestBuyRepository_DeductCoins_Error(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	queries := db.New(mockPool)
	logger := utils.NewLogger()
	repo := repository.NewBuyRepository(mockPool, queries, logger)

	ctx := context.Background()
	userID := int32(100)
	amount := int32(9999)

	mockPool.
		ExpectExec(regexp.QuoteMeta(`UPDATE employees SET coins = coins + $2 WHERE id = $1 AND coins + $2 >= 0`)).
		WithArgs(userID, -amount).
		WillReturnError(assert.AnError)

	affected, err := repo.DeductCoins(ctx, userID, amount)
	assert.Error(t, err)
	assert.Equal(t, int64(0), affected)

	assert.NoError(t, mockPool.ExpectationsWereMet())
}

// 8. UpsertInventory: успех.
func TestBuyRepository_UpsertInventory_Success(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	queries := db.New(mockPool)
	logger := utils.NewLogger()
	repo := repository.NewBuyRepository(mockPool, queries, logger)

	ctx := context.Background()
	params := db.UpsertInventoryParams{
		EmployeeID: 123,
		MerchID:    1,
		Quantity:   2,
	}

	// Допустим sqlc генерирует:
	// INSERT INTO inventory ...
	// ON CONFLICT (employee_id, merch_id) DO UPDATE ...
	mockPool.
		ExpectExec(regexp.QuoteMeta(`INSERT INTO inventory (employee_id, merch_id, quantity) VALUES ($1, $2, $3)
ON CONFLICT (employee_id, merch_id) DO UPDATE SET quantity = inventory.quantity + EXCLUDED.quantity`)).
		WithArgs(params.EmployeeID, params.MerchID, params.Quantity).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = repo.UpsertInventory(ctx, params)
	assert.NoError(t, err)

	assert.NoError(t, mockPool.ExpectationsWereMet())
}

// 9. UpsertInventory: ошибка.
func TestBuyRepository_UpsertInventory_Error(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	queries := db.New(mockPool)
	logger := utils.NewLogger()
	repo := repository.NewBuyRepository(mockPool, queries, logger)

	ctx := context.Background()
	params := db.UpsertInventoryParams{
		EmployeeID: 999,
		MerchID:    5,
		Quantity:   1,
	}

	mockPool.
		ExpectExec(regexp.QuoteMeta(`INSERT INTO inventory (employee_id, merch_id, quantity) VALUES ($1, $2, $3)
ON CONFLICT (employee_id, merch_id) DO UPDATE SET quantity = inventory.quantity + EXCLUDED.quantity`)).
		WithArgs(params.EmployeeID, params.MerchID, params.Quantity).
		WillReturnError(assert.AnError)

	err = repo.UpsertInventory(ctx, params)
	assert.Error(t, err)

	assert.NoError(t, mockPool.ExpectationsWereMet())
}

// 10. CreatePurchaseTransaction: успех.
func TestBuyRepository_CreatePurchaseTransaction_Success(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	queries := db.New(mockPool)
	logger := utils.NewLogger()
	repo := repository.NewBuyRepository(mockPool, queries, logger)

	ctx := context.Background()
	params := db.CreateCoinTransactionPurchaseParams{
		FromEmployeeID: 100,
		MerchID:        pgtype.Int4{Int32: 1, Valid: true},
		Amount:         50,
	}

	// sqlc генерирует что-то вроде:
	// INSERT INTO coin_transactions (transaction_type, from_employee_id, merch_id, amount)
	// VALUES ('purchase', $1, $2, $3)
	mockPool.
		ExpectExec(regexp.QuoteMeta(`INSERT INTO coin_transactions (transaction_type, from_employee_id, merch_id, amount) VALUES ('purchase', $1, $2, $3)`)).
		WithArgs(params.FromEmployeeID, params.MerchID, params.Amount).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = repo.CreatePurchaseTransaction(ctx, params)
	assert.NoError(t, err)

	assert.NoError(t, mockPool.ExpectationsWereMet())
}

// 11. CreatePurchaseTransaction: ошибка
func TestBuyRepository_CreatePurchaseTransaction_Error(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	queries := db.New(mockPool)
	logger := utils.NewLogger()
	repo := repository.NewBuyRepository(mockPool, queries, logger)

	ctx := context.Background()
	params := db.CreateCoinTransactionPurchaseParams{
		FromEmployeeID: 100,
		// merch_id ...
		Amount: 50,
	}

	mockPool.
		ExpectExec(regexp.QuoteMeta(`INSERT INTO coin_transactions (transaction_type, from_employee_id, merch_id, amount) VALUES ('purchase', $1, $2, $3)`)).
		WithArgs(params.FromEmployeeID, params.MerchID, params.Amount).
		WillReturnError(assert.AnError)

	err = repo.CreatePurchaseTransaction(ctx, params)
	assert.Error(t, err)

	assert.NoError(t, mockPool.ExpectationsWereMet())
}
