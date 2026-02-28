package config

import (
	"fmt"
	"os"
)

type Config struct {
	AppPort      string
	DBConnString string // Строка подключения к Postgres
	RedisAddr    string
}

func Load() *Config {
	// Собираем строку подключения из переменных окружения
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	return &Config{
		AppPort:      os.Getenv("APP_PORT"),
		DBConnString: dbURL,
		RedisAddr:    os.Getenv("REDIS_HOST"),
	}
}
