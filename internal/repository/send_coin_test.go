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

func TestSendCoinRepository_GetRecipient_Success(t *testing.T) {
	// Создаем моковый пул.
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	// Создаем экземпляр db.Queries с использованием мокового пула.
	queries := db.New(mockPool)

	now := time.Now() // или time.Now(), с учётом того, как вы его используете
	createdAt := pgtype.Timestamptz{Time: now, Valid: true}
	rows := pgxmock.NewRows([]string{"id", "username", "password_hash", "coins", "created_at"}).
		AddRow(int32(2), "recipient_user", "somehash", int32(200), createdAt)

	queryRegex := regexp.MustCompile("(?s)SELECT.*FROM employees.*WHERE username = \\$1")
	mockPool.ExpectQuery(queryRegex.String()).
		WithArgs("recipient_user").
		WillReturnRows(rows)

	// Создаем репозиторий через конструктор, который теперь принимает PoolIface.
	repoInstance := repository.NewSendCoinRepository(mockPool, queries, utils.NewLogger())
	recipient, err := repoInstance.GetRecipient(context.Background(), "recipient_user")
	assert.NoError(t, err)
	assert.NotNil(t, recipient)
	assert.Equal(t, int32(2), recipient.ID)
	assert.Equal(t, "recipient_user", recipient.Username)

	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestSendCoinRepository_GetBalance_Success(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	queries := db.New(mockPool)

	// Ожидаем запрос: SELECT coins FROM employees WHERE id=$1
	rows := pgxmock.NewRows([]string{"coins"}).AddRow(int32(150))
	queryRegex := regexp.MustCompile("(?s)SELECT coins FROM employees WHERE id=\\$1")
	mockPool.ExpectQuery(queryRegex.String()).
		WithArgs(int32(1)).
		WillReturnRows(rows)

	repoInstance := repository.NewSendCoinRepository(mockPool, queries, utils.NewLogger())
	balance, err := repoInstance.GetBalance(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, int32(150), balance)

	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestSendCoinRepository_TransferCoins_Success(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	queries := db.New(mockPool)
	repoInstance := repository.NewSendCoinRepository(mockPool, queries, utils.NewLogger())

	fromUserID := int32(1)
	toUserID := int32(2)
	amount := int32(50)

	// 1. Ожидаем вызов UpdateEmployeeCoins для списания средств у отправителя.
	updateQueryRegex := regexp.MustCompile("(?s)UPDATE employees\\s+SET coins = coins \\+ \\$2\\s+WHERE id = \\$1 AND coins \\+ \\$2 >= 0")
	mockPool.ExpectExec(updateQueryRegex.String()).
		WithArgs(fromUserID, -amount).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	// 2. Ожидаем вызов UpdateEmployeeCoins для начисления средств получателю.
	mockPool.ExpectExec(updateQueryRegex.String()).
		WithArgs(toUserID, amount).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	// 3. Ожидаем вызов CreateCoinTransactionTransfer для записи транзакции.
	ctQueryRegex := regexp.MustCompile("(?s)INSERT INTO coin_transactions .*VALUES \\('transfer', \\$1, \\$2, \\$3\\)")
	mockPool.ExpectExec(ctQueryRegex.String()).
		WithArgs(fromUserID, pgtype.Int4{Int32: toUserID, Valid: true}, amount).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = repoInstance.TransferCoins(context.Background(), fromUserID, toUserID, amount)
	assert.NoError(t, err)

	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestSendCoinRepository_ExecTx_Success(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	// Ожидаем начало транзакции.
	mockPool.ExpectBegin()
	// При успешном завершении транзакции ожидаем commit.
	mockPool.ExpectCommit()

	queries := db.New(mockPool)
	repoInstance := repository.NewSendCoinRepository(mockPool, queries, utils.NewLogger())

	err = repoInstance.ExecTx(context.Background(), func(repo repository.SendCoinRepository) error {
		// В callback можно вызывать методы репозитория, здесь достаточно вернуть nil.
		return nil
	})
	assert.NoError(t, err)

	assert.NoError(t, mockPool.ExpectationsWereMet())
}
