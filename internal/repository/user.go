package repository

import (
	"context"

	"github.com/par1ram/merch-store/internal/db"
)

// User представляет сотрудника.
type User struct {
	ID           int64  `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"password_hash"`
	Coins        int32  `json:"coins"`
}

type UserRepository interface {
	GetByUsername(ctx context.Context, username string) (User, error)
	Create(ctx context.Context, username, passwordHash string) (User, error)
}

type PostgresUserRepository struct {
	Queries *db.Queries
}

func NewPostgresUserRepository(q *db.Queries) UserRepository {
	return &PostgresUserRepository{
		Queries: q,
	}
}

// GetByUsername возвращает пользователя по username.
func (r *PostgresUserRepository) GetByUsername(ctx context.Context, username string) (User, error) {
	emp, err := r.Queries.GetEmployeeByUsername(ctx, username)
	if err != nil {
		return User{}, err
	}
	return User{
		ID:           int64(emp.ID),
		Username:     emp.Username,
		PasswordHash: emp.PasswordHash,
		Coins:        emp.Coins,
	}, nil
}

// Create создаёт нового пользователя.
func (r *PostgresUserRepository) Create(ctx context.Context, username, passwordHash string) (User, error) {
	emp, err := r.Queries.CreateEmployee(ctx, db.CreateEmployeeParams{
		Username:     username,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return User{}, err
	}
	return User{
		ID:           int64(emp.ID),
		Username:     emp.Username,
		PasswordHash: emp.PasswordHash,
		Coins:        emp.Coins,
	}, nil
}
