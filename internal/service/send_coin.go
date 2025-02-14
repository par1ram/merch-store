package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/par1ram/merch-store/internal/middleware"
	"github.com/par1ram/merch-store/internal/repository"
	"github.com/par1ram/merch-store/internal/utils"
)

var (
	ErrBusinessValidation = errors.New("business validation error")
	ErrSelfTransfer       = errors.New("self-transfer prohibited")
	ErrInsufficientFunds  = errors.New("insufficient funds")
	ErrRecipientNotFound  = errors.New("recipient not found")
)

type SendCoinService interface {
	SendCoin(ctx context.Context, toUser string, amount int32) error
}

type sendCoinService struct {
	repo   repository.SendCoinRepository
	logger utils.Logger
}

func NewSendCoinService(repo repository.SendCoinRepository, logger utils.Logger) SendCoinService {
	logger.WithFields(utils.LogFields{
		"component": "send_coin_service",
	}).Info("SendCoinService initialized")

	return &sendCoinService{
		repo:   repo,
		logger: logger,
	}
}

func (s *sendCoinService) SendCoin(ctx context.Context, toUser string, amount int32) error {
	log := s.logger.WithFields(utils.LogFields{
		"operation": "send_coin",
		"to_user":   toUser,
		"amount":    amount,
	})

	senderID := middleware.GetUserIDFromContext(ctx)
	if senderID == 0 {
		log.Error("Authentication required", utils.LogFields{
			"error": "missing_user_id",
		})
		return errors.New("authentication required")
	}

	log = log.WithFields(utils.LogFields{"from_user_id": senderID})
	log.Debug("Starting transaction")

	err := s.repo.ExecTx(ctx, func(r repository.SendCoinRepository) error {
		recipient, err := r.GetRecipient(ctx, toUser)
		if err != nil {
			log.Error("Recipient lookup failed", utils.LogFields{
				"error":      err,
				"error_type": "recipient_not_found",
				"recipient":  toUser,
			})
			return fmt.Errorf("%w: %v", ErrBusinessValidation, ErrRecipientNotFound)
		}

		if recipient.ID == int32(senderID) {
			log.Warn("Self-transfer attempt", utils.LogFields{
				"error_type":   "self_transfer",
				"sender_id":    senderID,
				"recipient_id": recipient.ID,
			})
			return fmt.Errorf("%w: %v", ErrBusinessValidation, ErrSelfTransfer)
		}

		balance, err := r.GetBalance(ctx, int32(senderID))
		if err != nil {
			log.Error("Balance check failed", utils.LogFields{
				"error":      err,
				"error_type": "balance_check_failed",
				"user_id":    senderID,
			})
			return fmt.Errorf("internal server error")
		}

		if balance < amount {
			log.Warn("Insufficient funds", utils.LogFields{
				"current_balance": balance,
				"required_amount": amount,
				"error_type":      "insufficient_funds",
			})
			return fmt.Errorf("%w: %v", ErrBusinessValidation, ErrInsufficientFunds)
		}

		if err := r.TransferCoins(ctx, int32(senderID), recipient.ID, amount); err != nil {
			log.Error("Transfer failed", utils.LogFields{
				"error":           err,
				"error_type":      "transfer_failure",
				"sender_id":       senderID,
				"recipient_id":    recipient.ID,
				"transfer_amount": amount,
			})
			return err
		}

		log.Info("Transfer successful", utils.LogFields{
			"sender_id":       senderID,
			"recipient_id":    recipient.ID,
			"transfer_amount": amount,
		})
		return nil
	})

	if err != nil {
		log.Error("Transaction failed", utils.LogFields{
			"error":      err,
			"error_type": "transaction_failure",
		})
		return err
	}

	log.Info("Transaction completed successfully")
	return nil
}
