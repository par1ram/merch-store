// file: repository/get_merch_test.go
package repository_test

import (
	"context"
	"regexp"
	"testing"

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
		AddRow(int32(1), "T-Shirt", int32(100)) // Используем int32 вместо int

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
