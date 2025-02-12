package main

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/par1ram/merch-store/internal/config"
	"github.com/par1ram/merch-store/internal/db"
	"github.com/par1ram/merch-store/internal/handlers"
	"github.com/par1ram/merch-store/internal/middleware"
	"github.com/par1ram/merch-store/internal/repository"
	"github.com/par1ram/merch-store/internal/service"
	"github.com/par1ram/merch-store/internal/utils"
	"github.com/pressly/goose"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := utils.NewLogger()
	cfg := config.LoadConfig()

	// Соединение для миграций
	sqlDB, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		logrus.Fatalf("Failed to open database: %v", err)
	}
	defer sqlDB.Close()

	// Применяем миграции
	if err := goose.Up(sqlDB, "internal/sql/schema"); err != nil {
		logrus.Fatalf("Error applying migrations: %v", err)
	}

	// Создание пула соединений.
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		logrus.Fatalf("unable to connect to database: %v", err)
	}
	defer pool.Close()

	// Seed: обогащаем базу данных данными о товарах.
	if err := utils.SeedMerchData(context.Background(), pool, logrus.StandardLogger()); err != nil {
		logrus.Fatalf("Error seeding merchandise data: %v", err)
	}

	// Инициализируем sqlc-клиент (сгенерированный код).
	queries := db.New(pool)

	userRepo := repository.NewPostgresUserRepository(queries)
	authService := service.NewAuthService(userRepo, []byte(cfg.JWTSecret))
	authHandler := handlers.NewAuthHandler(authService)

	infoRepo := repository.NewInfoRepository(queries, logger)
	infoService := service.NewInfoService(infoRepo, logger)
	infoHandler := handlers.NewInfoHandler(infoService)
	secureInfoHandler := middleware.JWTMiddleware([]byte(cfg.JWTSecret))(http.HandlerFunc(infoHandler.HandleInfo))

	sendCoinRepository := repository.NewSendCoinRepository(pool, queries, logger)
	sendCoinService := service.NewSendCoinService(sendCoinRepository, logger)
	sendCoinHandler := handlers.NewSendCoinHandler(sendCoinService)
	secureSendCoinHandler := middleware.JWTMiddleware([]byte(cfg.JWTSecret))(http.HandlerFunc(sendCoinHandler.HandleSendCoin))

	buyRepo := repository.NewBuyRepository(pool, queries, logger)
	buyService := service.NewBuyService(buyRepo, logger)
	buyHandler := handlers.NewBuyHandler(buyService)
	secureBuyHandler := middleware.JWTMiddleware([]byte(cfg.JWTSecret))(http.HandlerFunc(buyHandler.HandleBuy))

	// Маршруты.
	mux := http.NewServeMux()
	mux.HandleFunc("/api/auth", authHandler.HandleAuth)
	mux.Handle("/api/info", secureInfoHandler)
	mux.Handle("/api/send-coin", secureSendCoinHandler)
	mux.Handle("/api/buy/", secureBuyHandler)

	logrus.Info("Server started on PORT: ", cfg.ServerPort)
	if err := http.ListenAndServe(":"+cfg.ServerPort, mux); err != nil {
		logrus.Fatalf("failed to start server: %v", err)
	}
}
