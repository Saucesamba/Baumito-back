package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"Avito-back/internal/config"
	"Avito-back/internal/repository/postgres"
	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Загружаем конфиг
	cfg := config.Load()

	// 2. Подключаемся к БД
	db, err := postgres.NewConnection(cfg.DBConnString)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// 3. Настраиваем роутер (Gin)
	router := gin.Default()

	// Простейший Health Check для проверки работы
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "Campus Marketplace is running!"})
	})

	// 4. Запуск сервера с Graceful Shutdown (красивое завершение работы)
	srv := &http.Server{
		Addr:    ":" + os.Getenv("APP_PORT"),
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	log.Printf("Server started on port %s", os.Getenv("APP_PORT"))

	// Ожидаем сигнал прерывания (Ctrl+C или Docker stop)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
