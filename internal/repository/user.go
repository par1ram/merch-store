package repository

import (
	"context"

	"github.com/par1ram/merch-store/internal/db"
	"github.com/par1ram/merch-store/internal/utils"
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
	logger  utils.Logger
}

// NewPostgresUserRepository создаёт новый репозиторий, принимающий sqlc‑клиент и логгер.
func NewPostgresUserRepository(q *db.Queries, logger utils.Logger) UserRepository {
	logger.WithFields(utils.LogFields{"component": "user_repository"}).Info("UserRepository initialized")
	return &PostgresUserRepository{
		Queries: q,
		logger:  logger.WithFields(utils.LogFields{"component": "postgres_user_repository"}),
	}
}

// GetByUsername возвращает пользователя по username.
func (r *PostgresUserRepository) GetByUsername(ctx context.Context, username string) (User, error) {
	r.logger.Infof("Getting user by username: %s", username)
	emp, err := r.Queries.GetEmployeeByUsername(ctx, username)
	if err != nil {
		r.logger.WithFields(utils.LogFields{
			"username": username,
			"error":    err,
		}).Errorf("Failed to get employee by username")
		return User{}, err
	}
	r.logger.WithFields(utils.LogFields{
		"userID":   emp.ID,
		"username": emp.Username,
	}).Debugf("User retrieved successfully")
	return User{
		ID:           int64(emp.ID),
		Username:     emp.Username,
		PasswordHash: emp.PasswordHash,
		Coins:        emp.Coins,
	}, nil
}

// Create создаёт нового пользователя.
func (r *PostgresUserRepository) Create(ctx context.Context, username, passwordHash string) (User, error) {
	r.logger.Infof("Creating new user: %s", username)
	emp, err := r.Queries.CreateEmployee(ctx, db.CreateEmployeeParams{
		Username:     username,
		PasswordHash: passwordHash,
	})
	if err != nil {
		r.logger.WithFields(utils.LogFields{
			"username": username,
			"error":    err,
		}).Errorf("Failed to create employee")
		return User{}, err
	}
	r.logger.WithFields(utils.LogFields{
		"userID":   emp.ID,
		"username": emp.Username,
	}).Debugf("User created successfully")
	return User{
		ID:           int64(emp.ID),
		Username:     emp.Username,
		PasswordHash: emp.PasswordHash,
		Coins:        emp.Coins,
	}, nil
}
