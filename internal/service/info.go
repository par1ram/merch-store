package service

import (
	"context"

	"github.com/par1ram/merch-store/internal/repository"
	"github.com/par1ram/merch-store/internal/utils"
)

type InfoResponse struct {
	Coins       int         `json:"coins"`
	Inventory   []Inventory `json:"inventory"`
	CoinHistory CoinHistory `json:"coinHistory"`
}

type Inventory struct {
	Type     string `json:"type"`
	Quantity int    `json:"quantity"`
}

type CoinHistory struct {
	Received []ReceivedTransaction `json:"received"`
	Sent     []SentTransaction     `json:"sent"`
}

type ReceivedTransaction struct {
	FromUser string `json:"fromUser"`
	Amount   int    `json:"amount"`
}

type SentTransaction struct {
	ToUser string `json:"toUser"`
	Amount int    `json:"amount"`
}

type InfoService interface {
	GetInfo(ctx context.Context, userID int64) (InfoResponse, error)
}

type infoService struct {
	repo   repository.InfoRepository
	logger utils.Logger
}

func NewInfoService(repo repository.InfoRepository, logger utils.Logger) InfoService {
	logger.WithFields(utils.LogFields{"component": "info_service"}).Info("InfoService initialized")
	return &infoService{
		repo:   repo,
		logger: logger.WithFields(utils.LogFields{"component": "info_service"}),
	}
}

func (s *infoService) GetInfo(ctx context.Context, userID int64) (InfoResponse, error) {
	log := s.logger.WithFields(utils.LogFields{
		"operation": "get_info",
		"user_id":   userID,
	})

	log.Debug("Starting info retrieval")

	coins, err := s.repo.GetCoins(ctx, userID)
	if err != nil {
		log.WithFields(utils.LogFields{"error": err}).Error("Failed to get coins")
		return InfoResponse{}, err
	}
	log.WithFields(utils.LogFields{"coins": coins}).Debug("Coins retrieved")

	invItems, err := s.repo.GetInventory(ctx, userID)
	if err != nil {
		log.WithFields(utils.LogFields{"error": err}).Error("Failed to get inventory")
		return InfoResponse{}, err
	}
	log.WithFields(utils.LogFields{"inventory_count": len(invItems)}).Debug("Inventory retrieved")

	inventory := make([]Inventory, 0, len(invItems))
	for _, item := range invItems {
		inventory = append(inventory, Inventory{
			Type:     item.Type,
			Quantity: item.Quantity,
		})
	}

	rec, err := s.repo.GetReceivedTransfers(ctx, userID)
	if err != nil {
		log.WithFields(utils.LogFields{"error": err}).Error("Failed to get received transfers")
		return InfoResponse{}, err
	}
	log.WithFields(utils.LogFields{"received_count": len(rec)}).Debug("Received transfers retrieved")

	recTrans := make([]ReceivedTransaction, 0, len(rec))
	for _, t := range rec {
		recTrans = append(recTrans, ReceivedTransaction{
			FromUser: t.FromUser,
			Amount:   t.Amount,
		})
	}

	sent, err := s.repo.GetSentTransfers(ctx, userID)
	if err != nil {
		log.WithFields(utils.LogFields{"error": err}).Error("Failed to get sent transfers")
		return InfoResponse{}, err
	}
	log.WithFields(utils.LogFields{"sent_count": len(sent)}).Debug("Sent transfers retrieved")

	sentTrans := make([]SentTransaction, 0, len(sent))
	for _, t := range sent {
		sentTrans = append(sentTrans, SentTransaction{
			ToUser: t.ToUser,
			Amount: t.Amount,
		})
	}

	response := InfoResponse{
		Coins:     coins,
		Inventory: inventory,
		CoinHistory: CoinHistory{
			Received: recTrans,
			Sent:     sentTrans,
		},
	}

	log.WithFields(utils.LogFields{
		"total_coins":        coins,
		"inventory_items":    len(inventory),
		"received_transfers": len(recTrans),
		"sent_transfers":     len(sentTrans),
	}).Info("Info retrieved successfully")

	return response, nil
}
