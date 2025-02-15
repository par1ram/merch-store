package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"

	"github.com/par1ram/merch-store/internal/db"
	"github.com/par1ram/merch-store/internal/repository"
	"github.com/par1ram/merch-store/internal/utils"
)

func TestUserRepository_GetByUsername_Success(t *testing.T) {
	// Создаём моковый пул pgxmock.
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	// Создаём sqlc.Queries, используя мок‑пул.
	queries := db.New(mockPool)
	logger := utils.NewLogger()
	userRepo := repository.NewPostgresUserRepository(queries, logger)

	ctx := context.Background()
	username := "alice"

	// Заполним Timestamptz вручную.
	var createdAt pgtype.Timestamptz
	createdAt.Time = time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	createdAt.Valid = true

	// Подготавливаем строку с SQL (полностью или частично).
	// Для простоты предположим, что sqlc генерирует именно такой запрос:
	// "SELECT id, username, password_hash, coins, created_at FROM employees WHERE username = $1"
	mockPool.
		ExpectQuery(`SELECT id, username, password_hash, coins, created_at FROM employees WHERE username = \$1`).
		WithArgs(username).
		WillReturnRows(
			mockPool.NewRows([]string{"id", "username", "password_hash", "coins", "created_at"}).
				AddRow(int32(10), "alice", "hash123", int32(100), createdAt),
		)

	// Вызываем метод
	user, err := userRepo.GetByUsername(ctx, username)
	assert.NoError(t, err)
	assert.Equal(t, int64(10), user.ID)
	assert.Equal(t, "alice", user.Username)
	assert.Equal(t, "hash123", user.PasswordHash)
	assert.Equal(t, int32(100), user.Coins)

	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestUserRepository_GetByUsername_NoRows(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	queries := db.New(mockPool)
	logger := utils.NewLogger()
	userRepo := repository.NewPostgresUserRepository(queries, logger)

	ctx := context.Background()
	username := "nonexistent"

	// Настраиваем пустой результат, чтобы вернулось sql.ErrNoRows.
	mockPool.
		ExpectQuery(`SELECT id, username, password_hash, coins, created_at FROM employees WHERE username = \$1`).
		WithArgs(username).
		WillReturnRows(mockPool.NewRows([]string{"id", "username", "password_hash", "coins", "created_at"})) // без строк

	// Вызываем метод
	user, err := userRepo.GetByUsername(ctx, username)
	assert.Error(t, err)
	assert.Empty(t, user)

	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestUserRepository_Create_Success(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	queries := db.New(mockPool)
	logger := utils.NewLogger()
	userRepo := repository.NewPostgresUserRepository(queries, logger)

	ctx := context.Background()
	username := "newuser"
	passwordHash := "hashedpass"

	mockPool.
		ExpectQuery(`INSERT INTO employees.*RETURNING id, username, coins, password_hash`).
		WithArgs(username, passwordHash).
		WillReturnRows(
			mockPool.NewRows([]string{"id", "username", "coins", "password_hash"}).
				AddRow(int32(11), username, int32(0), passwordHash),
		)

	user, err := userRepo.Create(ctx, username, passwordHash)
	assert.NoError(t, err)
	assert.Equal(t, int64(11), user.ID)
	assert.Equal(t, "newuser", user.Username)
	assert.Equal(t, "hashedpass", user.PasswordHash)
	assert.Equal(t, int32(0), user.Coins)

	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestUserRepository_Create_Error(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	queries := db.New(mockPool)
	logger := utils.NewLogger()
	userRepo := repository.NewPostgresUserRepository(queries, logger)

	ctx := context.Background()
	username := "failuser"
	passwordHash := "somehash"

	// Ставим те же самые поля, что в реальном запросе:
	mockPool.
		ExpectQuery(`INSERT INTO employees.*RETURNING id, username, coins, password_hash`).
		WithArgs(username, passwordHash).
		WillReturnError(assert.AnError) // Имитация ошибки при вставке.

	user, err := userRepo.Create(ctx, username, passwordHash)

	assert.Error(t, err)  // Ожидаем ошибку
	assert.Empty(t, user) // Пустой пользователь
	assert.NoError(t, mockPool.ExpectationsWereMet())
}
