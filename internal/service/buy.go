package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/par1ram/merch-store/internal/db"
	"github.com/par1ram/merch-store/internal/middleware"
	"github.com/par1ram/merch-store/internal/repository"
	"github.com/par1ram/merch-store/internal/utils"
)

// BuyService определяет метод покупки товара.
type BuyService interface {
	Purchase(ctx context.Context, item string) error
}

type buyService struct {
	repo   repository.BuyRepository
	logger utils.Logger
}

func NewBuyService(repo repository.BuyRepository, logger utils.Logger) BuyService {
	logger.WithFields(map[string]interface{}{"component": "buy_service"}).Info("BuyService initialized")
	return &buyService{
		repo:   repo,
		logger: logger,
	}
}

func (s *buyService) Purchase(ctx context.Context, item string) error {
	// Извлекаем идентификатор пользователя из контекста.
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == 0 {
		s.logger.Error("User not authenticated")
		return errors.New("user not authenticated")
	}
	s.logger.Infof("Processing purchase; userID=%d, item=%s", userID, item)

	// Получаем информацию о товаре.
	merch, err := s.repo.GetMerch(ctx, item)
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"error":   err,
			"item":    item,
			"user_id": userID,
		}).Error("Failed to get merch")
		return fmt.Errorf("merch not found: %w", err)
	}
	s.logger.Infof("Merch found; item=%s, price=%d", merch.Name, merch.Price)

	// Проверяем, достаточно ли средств у пользователя.
	balance, err := s.repo.GetBalance(ctx, int32(userID))
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"error":   err,
			"user_id": userID,
		}).Error("Failed to get user balance")
		return fmt.Errorf("failed to get balance: %w", err)
	}
	if balance < merch.Price {
		s.logger.Warnf("Insufficient funds; userID=%d, balance=%d, price=%d", userID, balance, merch.Price)
		return errors.New("insufficient funds")
	}

	// Запускаем транзакцию для покупки товара.
	err = s.repo.ExecTx(ctx, func(r repository.BuyRepository) error {
		// Списываем монеты с баланса пользователя.
		affected, err := r.DeductCoins(ctx, int32(userID), merch.Price)
		if err != nil {
			return err
		}
		if affected == 0 {
			return errors.New("failed to deduct coins")
		}

		// Обновляем инвентарь: добавляем единицу купленного товара.
		upsertParams := db.UpsertInventoryParams{
			EmployeeID: int32(userID),
			MerchID:    merch.ID,
			Quantity:   1,
		}
		if err := r.UpsertInventory(ctx, upsertParams); err != nil {
			return err
		}

		// Регистрируем транзакцию покупки.
		purchaseParams := db.CreateCoinTransactionPurchaseParams{
			FromEmployeeID: int32(userID),
			MerchID:        pgtype.Int4{Int32: merch.ID, Valid: true},
			Amount:         merch.Price,
		}
		if err := r.CreatePurchaseTransaction(ctx, purchaseParams); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"error":   err,
			"user_id": userID,
			"item":    item,
		}).Error("Purchase transaction failed")
		return err
	}

	s.logger.Infof("Purchase successful; userID=%d, item=%s, price=%d", userID, merch.Name, merch.Price)
	return nil
}
