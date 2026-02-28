package postgres

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Storage структура для хранения пула соединений
type Storage struct {
	Pool *pgxpool.Pool
}

// NewConnection создает пул соединений к БД
func NewConnection(connString string) (*Storage, error) {
	ctx := context.Background()
	var pool *pgxpool.Pool
	var err error

	// Пробуем подключиться 10 раз, так как БД в докере может запускаться долго
	for i := 0; i < 10; i++ {
		pool, err = pgxpool.New(ctx, connString)
		if err == nil {
			err = pool.Ping(ctx)
			if err == nil {
				log.Println("Successfully connected to PostgreSQL!")
				return &Storage{Pool: pool}, nil
			}
		}

		log.Printf("DB not ready, retrying in 2 seconds... (attempt %d/10)", i+1)
		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("failed to connect to postgres after 10 attempts: %w", err)
}

func (s *Storage) Close() {
	s.Pool.Close()
}
