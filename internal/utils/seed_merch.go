package utils

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// SeedMerchData вставляет товары (merch) — как уже было.
// А также вставляет двух пользователей.
func SeedMerchData(ctx context.Context, pool *pgxpool.Pool, logger *logrus.Logger) error {
	items := []struct {
		Name  string
		Price int32
	}{
		{"t-shirt", 80},
		{"cup", 20},
		{"book", 50},
		{"pen", 10},
		{"powerbank", 200},
		{"hoody", 300},
		{"umbrella", 200},
		{"socks", 10},
		{"wallet", 50},
		{"pink-hoody", 500},
	}

	for _, item := range items {
		_, err := pool.Exec(ctx, `
			INSERT INTO merch (name, price)
			VALUES ($1, $2)
			ON CONFLICT (name) DO NOTHING
		`, item.Name, item.Price)
		if err != nil {
			logger.Errorf("Failed to insert merch '%s': %v", item.Name, err)
			return fmt.Errorf("failed to insert merch '%s': %w", item.Name, err)
		}
		logger.Infof("Merch '%s' inserted (or already exists)", item.Name)
	}

	users := []struct {
		Username string
		Password string
		Coins    int32
	}{
		{"testuser", "pass123", 1000},
		{"alice", "pass456", 1000},
	}

	for _, u := range users {
		// Генерируем bcrypt-хэш
		hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			logger.Errorf("Failed to bcrypt password for user '%s': %v", u.Username, err)
			return fmt.Errorf("failed to bcrypt password for user '%s': %w", u.Username, err)
		}

		// Вставляем пользователя
		_, err = pool.Exec(ctx, `
			INSERT INTO employees (username, password_hash, coins)
			VALUES ($1, $2, $3)
			ON CONFLICT (username) DO NOTHING
		`, u.Username, string(hash), u.Coins)
		if err != nil {
			logger.Errorf("Failed to insert user '%s': %v", u.Username, err)
			return fmt.Errorf("failed to insert user '%s': %w", u.Username, err)
		}
		logger.Infof("User '%s' inserted (or already exists)", u.Username)
	}

	return nil
}
