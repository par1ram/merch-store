package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config структура для хранения настроек приложения
type Config struct {
	ServerPort  string
	DatabaseURL string
	JWTSecret   string
}

// LoadConfig загружает конфигурацию из .env или переменных окружения
func LoadConfig() *Config {
	// Загружаем .env файл, если он существует
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, loading from environment variables")
	}

	return &Config{
		ServerPort:  getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://user:password@localhost:5432/dbname"),
		JWTSecret:   getEnv("JWT_SECRET", "your-secret-key"),
	}
}

// getEnv возвращает значение переменной окружения или значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
