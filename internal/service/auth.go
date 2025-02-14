package service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx"
	"github.com/par1ram/merch-store/internal/repository"
	"github.com/par1ram/merch-store/internal/utils"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

// AuthService определяет интерфейс для аутентификации.
type AuthService interface {
	Authenticate(ctx context.Context, username, password string) (string, error)
}

// authService — конкретная реализация AuthService.
type authService struct {
	userRepo  repository.UserRepository
	jwtSecret []byte
	logger    utils.Logger
}

// NewAuthService создаёт новый AuthService, используя репозиторий и логгер.
func NewAuthService(userRepo repository.UserRepository, jwtSecret []byte, logger utils.Logger) AuthService {
	logger.WithFields(utils.LogFields{
		"component": "auth_service",
	}).Info("AuthService initialized")
	return &authService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
		logger:    logger,
	}
}

// Authenticate проверяет учётные данные пользователя и возвращает JWT-токен.
func (s *authService) Authenticate(ctx context.Context, username, password string) (string, error) {
	s.logger.Infof("Authenticating user: %s", username)

	// Попытка найти пользователя.
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		// Если пользователь не найден, создаём нового.
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			s.logger.Infof("User %s not found, creating new user", username)
			hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				s.logger.Errorf("Error generating password hash for user %s: %v", username, err)
				return "", err
			}
			user, err = s.userRepo.Create(ctx, username, string(hash))
			if err != nil {
				s.logger.Errorf("Error creating user %s: %v", username, err)
				return "", err
			}
			s.logger.Infof("User %s created successfully, id: %d", username, user.ID)
		} else {
			s.logger.Errorf("Error retrieving user %s: %v", username, err)
			return "", err
		}
	} else {
		// Пользователь найден — сравниваем хэш пароля.
		s.logger.Infof("User %s found, verifying password", username)
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
			s.logger.Errorf("Invalid credentials for user %s: %v", username, err)
			return "", errors.New("invalid credentials")
		}
		s.logger.Infof("Password verification succeeded for user %s", username)
	}

	// Генерируем JWT-токен.
	token, err := generateJWT(user, s.jwtSecret)
	if err != nil {
		s.logger.Errorf("Error generating JWT for user %s: %v", username, err)
		return "", err
	}
	s.logger.Infof("JWT generated successfully for user %s", username)
	return token, nil
}

// generateJWT создаёт JWT-токен с информацией о пользователе.
func generateJWT(user repository.User, jwtSecret []byte) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      expirationTime.Unix(),
		"iat":      time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}
