package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/par1ram/merch-store/internal/db"
	"github.com/par1ram/merch-store/internal/utils"
)

type BuyRepository interface {
	ExecTx(ctx context.Context, fn func(BuyRepository) error) error
	GetMerch(ctx context.Context, merchName string) (db.Merch, error)
	GetBalance(ctx context.Context, userID int32) (int32, error)
	DeductCoins(ctx context.Context, userID, amount int32) (int64, error)
	UpsertInventory(ctx context.Context, params db.UpsertInventoryParams) error
	CreatePurchaseTransaction(ctx context.Context, params db.CreateCoinTransactionPurchaseParams) error
}

type buyRepository struct {
	pool    *pgxpool.Pool
	queries *db.Queries
	logger  utils.Logger
}

func NewBuyRepository(pool *pgxpool.Pool, queries *db.Queries, logger utils.Logger) BuyRepository {
	logger.WithFields(utils.LogFields{"component": "buy_repository"}).Info("BuyRepository initialized")
	return &buyRepository{
		pool:    pool,
		queries: queries,
		logger:  logger,
	}
}

func (r *buyRepository) ExecTx(ctx context.Context, fn func(BuyRepository) error) error {
	r.logger.WithFields(utils.LogFields{"operation": "exec_tx"}).Info("Starting transaction")
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		r.logger.WithFields(utils.LogFields{"error": err}).Error("Transaction begin failed")
		return fmt.Errorf("transaction start failed: %w", err)
	}
	defer tx.Rollback(ctx)

	// Получаем новый экземпляр queries, ассоциированный с транзакцией.
	qtx := r.queries.WithTx(tx)
	txRepo := &buyRepository{
		pool:    r.pool, // пул остаётся тем же
		queries: qtx,
		logger:  r.logger,
	}

	if err := fn(txRepo); err != nil {
		r.logger.WithFields(utils.LogFields{"error": err}).Error("Transaction operation failed")
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		r.logger.WithFields(utils.LogFields{"error": err}).Error("Transaction commit failed")
		return fmt.Errorf("transaction commit failed: %w", err)
	}

	r.logger.Info("Transaction committed")
	return nil
}

func (r *buyRepository) GetMerch(ctx context.Context, merchName string) (db.Merch, error) {
	merch, err := r.queries.GetMerchByName(ctx, merchName)
	if err != nil {
		r.logger.WithFields(utils.LogFields{"error": err, "merchName": merchName}).Error("Failed to get merch by name")
		return merch, err
	}
	r.logger.WithFields(utils.LogFields{
		"merchID":   merch.ID,
		"merchName": merch.Name,
		"price":     merch.Price,
	}).Debug("Merch found")
	return merch, nil
}

func (r *buyRepository) GetBalance(ctx context.Context, userID int32) (int32, error) {
	balance, err := r.queries.GetCoinsByID(ctx, userID)
	if err != nil {
		r.logger.WithFields(utils.LogFields{"error": err, "userID": userID}).Error("Failed to get balance")
		return 0, fmt.Errorf("get balance failed: %w", err)
	}
	r.logger.WithFields(utils.LogFields{"userID": userID, "balance": balance}).Debug("Balance retrieved")
	return balance, nil
}

func (r *buyRepository) DeductCoins(ctx context.Context, userID, amount int32) (int64, error) {
	err := r.queries.UpdateEmployeeCoins(ctx, db.UpdateEmployeeCoinsParams{
		ID:    userID,
		Coins: -amount,
	})
	if err != nil {
		r.logger.WithFields(utils.LogFields{"error": err, "userID": userID, "amount": amount}).Error("Failed to deduct coins")
		return 0, err
	}
	return 1, nil
}

func (r *buyRepository) UpsertInventory(ctx context.Context, params db.UpsertInventoryParams) error {
	if err := r.queries.UpsertInventory(ctx, params); err != nil {
		r.logger.WithFields(utils.LogFields{"error": err, "employee_id": params.EmployeeID, "merch_id": params.MerchID}).Error("Failed to upsert inventory")
		return err
	}
	return nil
}

func (r *buyRepository) CreatePurchaseTransaction(ctx context.Context, params db.CreateCoinTransactionPurchaseParams) error {
	if err := r.queries.CreateCoinTransactionPurchase(ctx, params); err != nil {
		r.logger.WithFields(utils.LogFields{"error": err, "fromUserID": params.FromEmployeeID, "merchID": params.MerchID}).Error("Failed to create purchase transaction")
		return err
	}
	return nil
}
