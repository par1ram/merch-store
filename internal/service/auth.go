package service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx"
	"github.com/par1ram/merch-store/internal/repository"

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
}

// NewAuthService создаёт новый AuthService.
func NewAuthService(userRepo repository.UserRepository, jwtSecret []byte) AuthService {
	return &authService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

// Authenticate проверяет учётные данные пользователя и возвращает JWT-токен.
func (s *authService) Authenticate(ctx context.Context, username, password string) (string, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				return "", err
			}
			user, err = s.userRepo.Create(ctx, username, string(hash))
			if err != nil {
				return "", err
			}
		} else {
			return "", err
		}
	} else {
		// Пользователь найден – сравниваем хэш пароля.
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
			return "", errors.New("invalid credentials")
		}
	}

	// Генерируем JWT-токен.
	token, err := generateJWT(user, s.jwtSecret)
	if err != nil {
		return "", err
	}
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
