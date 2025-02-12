package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/par1ram/merch-store/internal/db"
	"github.com/par1ram/merch-store/internal/utils"
)

type SendCoinRepository interface {
	ExecTx(ctx context.Context, fn func(SendCoinRepository) error) error
	GetRecipient(ctx context.Context, username string) (*db.Employee, error)
	GetBalance(ctx context.Context, userID int32) (int32, error)
	TransferCoins(ctx context.Context, fromUserID, toUserID, amount int32) error
}

type sendCoinRepository struct {
	pool    *pgxpool.Pool
	queries *db.Queries
	logger  utils.Logger
}

func NewSendCoinRepository(pool *pgxpool.Pool, queries *db.Queries, logger utils.Logger) SendCoinRepository {
	return &sendCoinRepository{
		pool:    pool,
		queries: queries,
		logger:  logger.WithFields(utils.LogFields{"component": "send_coin_repository"}),
	}
}

func (r *sendCoinRepository) ExecTx(ctx context.Context, fn func(SendCoinRepository) error) error {
	log := r.logger.WithFields(utils.LogFields{"operation": "exec_tx"})

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		log.WithFields(utils.LogFields{"error": err}).Error("transaction begin failed")
		return fmt.Errorf("transaction start failed: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)
	txRepo := &sendCoinRepository{
		queries: qtx,
		logger:  r.logger,
	}

	if err := fn(txRepo); err != nil {
		log.WithFields(utils.LogFields{"error": err}).Error("transaction operation failed")
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		log.WithFields(utils.LogFields{"error": err}).Error("transaction commit failed")
		return fmt.Errorf("transaction commit failed: %w", err)
	}

	log.Info("transaction committed")
	return nil
}

func (r *sendCoinRepository) GetRecipient(ctx context.Context, username string) (*db.Employee, error) {
	log := r.logger.WithFields(utils.LogFields{
		"operation": "get_recipient",
		"username":  username,
	})

	employee, err := r.queries.GetEmployeeByUsername(ctx, username)
	if err != nil {
		log.WithFields(utils.LogFields{"error": err}).Error("recipient lookup failed")
		return nil, fmt.Errorf("get recipient failed: %w", err)
	}

	log.WithFields(utils.LogFields{"user_id": employee.ID}).Debug("recipient found")
	return &employee, nil
}

func (r *sendCoinRepository) GetBalance(ctx context.Context, userID int32) (int32, error) {
	log := r.logger.WithFields(utils.LogFields{
		"operation": "get_balance",
		"user_id":   userID,
	})

	balance, err := r.queries.GetCoinsByID(ctx, userID)
	if err != nil {
		log.WithFields(utils.LogFields{"error": err}).Error("balance check failed")
		return 0, fmt.Errorf("get balance failed: %w", err)
	}

	log.WithFields(utils.LogFields{"balance": balance}).Debug("balance retrieved")
	return balance, nil
}

func (r *sendCoinRepository) TransferCoins(ctx context.Context, fromUserID, toUserID, amount int32) error {
	log := r.logger.WithFields(utils.LogFields{
		"operation":    "transfer_coins",
		"from_user_id": fromUserID,
		"to_user_id":   toUserID,
		"amount":       amount,
	})

	if err := r.queries.UpdateEmployeeCoins(ctx, db.UpdateEmployeeCoinsParams{
		ID:    fromUserID,
		Coins: -amount,
	}); err != nil {
		log.WithFields(utils.LogFields{"error": err}).Error("withdrawal failed")
		return fmt.Errorf("withdrawal failed: %w", err)
	}

	if err := r.queries.UpdateEmployeeCoins(ctx, db.UpdateEmployeeCoinsParams{
		ID:    toUserID,
		Coins: amount,
	}); err != nil {
		log.WithFields(utils.LogFields{"error": err}).Error("deposit failed")
		return fmt.Errorf("deposit failed: %w", err)
	}

	if err := r.queries.CreateCoinTransactionTransfer(ctx, db.CreateCoinTransactionTransferParams{
		FromEmployeeID: fromUserID,
		ToEmployeeID:   pgtype.Int4{Int32: toUserID, Valid: true},
		Amount:         amount,
	}); err != nil {
		log.WithFields(utils.LogFields{"error": err}).Error("transaction record creation failed")
		return fmt.Errorf("transaction record failed: %w", err)
	}

	log.Info("transfer completed")
	return nil
}
