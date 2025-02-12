package utils

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

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
	return nil
}
