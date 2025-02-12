package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/par1ram/merch-store/internal/db"
	"github.com/par1ram/merch-store/internal/utils"
)

type InventoryItem struct {
	Type     string
	Quantity int
}

type ReceivedTransaction struct {
	FromUser string
	Amount   int
}

type SentTransaction struct {
	ToUser string
	Amount int
}

type InfoRepository interface {
	GetCoins(ctx context.Context, userID int64) (int, error)
	GetInventory(ctx context.Context, userID int64) ([]InventoryItem, error)
	GetReceivedTransfers(ctx context.Context, userID int64) ([]ReceivedTransaction, error)
	GetSentTransfers(ctx context.Context, userID int64) ([]SentTransaction, error)
}

type infoRepository struct {
	Queries *db.Queries
	logger  utils.Logger
}

func NewInfoRepository(queries *db.Queries, logger utils.Logger) InfoRepository {
	return &infoRepository{
		Queries: queries,
		logger: logger.WithFields(utils.LogFields{
			"component": "info_repository",
		}),
	}
}

func (r *infoRepository) logOperation(ctx context.Context, operation string) utils.Logger {
	return r.logger.WithFields(utils.LogFields{
		"operation": operation,
		"user_id":   ctx.Value("user_id"),
	})
}

func (r *infoRepository) GetCoins(ctx context.Context, userID int64) (int, error) {
	log := r.logOperation(ctx, "get_coins")
	log.Debug("Starting coins retrieval")

	emp, err := r.Queries.GetEmployeeByID(ctx, int32(userID))
	if err != nil {
		log.WithFields(utils.LogFields{"error": err}).Error("Failed to get employee coins")
		return 0, err
	}

	log.WithFields(utils.LogFields{"coins": emp.Coins}).Debug("Coins retrieved successfully")
	return int(emp.Coins), nil
}

func (r *infoRepository) GetInventory(ctx context.Context, userID int64) ([]InventoryItem, error) {
	log := r.logOperation(ctx, "get_inventory")
	log.Debug("Starting inventory retrieval")

	items, err := r.Queries.GetInventoryByEmployeeID(ctx, int32(userID))
	if err != nil {
		log.WithFields(utils.LogFields{"error": err}).Error("Failed to get inventory")
		return nil, err
	}

	inventory := make([]InventoryItem, 0, len(items))
	for _, item := range items {
		inventory = append(inventory, InventoryItem{
			Type:     item.MerchName,
			Quantity: int(item.Quantity),
		})
	}

	log.WithFields(utils.LogFields{"item_count": len(inventory)}).Debug("Inventory retrieved")
	return inventory, nil
}

func (r *infoRepository) GetReceivedTransfers(ctx context.Context, userID int64) ([]ReceivedTransaction, error) {
	log := r.logOperation(ctx, "get_received_transfers")
	log.Debug("Starting received transfers retrieval")

	transfers, err := r.Queries.GetReceivedTransfers(ctx, pgtype.Int4{
		Int32: int32(userID),
		Valid: true,
	})
	if err != nil {
		log.WithFields(utils.LogFields{"error": err}).Error("Failed to get received transfers")
		return nil, err
	}

	recTrans := make([]ReceivedTransaction, 0, len(transfers))
	for _, t := range transfers {
		recTrans = append(recTrans, ReceivedTransaction{
			FromUser: t.FromUser,
			Amount:   int(t.Amount),
		})
	}

	log.WithFields(utils.LogFields{"transfer_count": len(recTrans)}).Debug("Received transfers retrieved")
	return recTrans, nil
}

func (r *infoRepository) GetSentTransfers(ctx context.Context, userID int64) ([]SentTransaction, error) {
	log := r.logOperation(ctx, "get_sent_transfers")
	log.Debug("Starting sent transfers retrieval")

	transfers, err := r.Queries.GetSentTransfers(ctx, int32(userID))
	if err != nil {
		log.WithFields(utils.LogFields{"error": err}).Error("Failed to get sent transfers")
		return nil, err
	}

	sentTrans := make([]SentTransaction, 0, len(transfers))
	for _, t := range transfers {
		sentTrans = append(sentTrans, SentTransaction{
			ToUser: t.ToUser,
			Amount: int(t.Amount),
		})
	}

	log.WithFields(utils.LogFields{"transfer_count": len(sentTrans)}).Debug("Sent transfers retrieved")
	return sentTrans, nil
}
