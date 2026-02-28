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
	"Avito-back/internal/delivery/http/v1"
	"Avito-back/internal/repository/postgres"
	"Avito-back/internal/usecase/user"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Загружаем конфигурацию
	cfg := config.Load()

	// 2. Инициализируем подключение к PostgreSQL (с логикой Retry внутри)
	db, err := postgres.NewConnection(cfg.DBConnString)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// 3. Инициализация слоев (Dependency Injection)

	// Repository (Слой работы с БД)
	userRepo := postgres.NewUserRepository(db)

	// Usecase (Слой бизнес-логики)
	userUsecase := user.NewUserUsecase(userRepo)

	// Delivery/Handler (Слой HTTP интерфейса)
	userHandler := &v1.UserHandler{
		Usecase: userUsecase,
	}

	// 4. Настройка роутера Gin
	router := gin.Default()

	// Базовый Health Check
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "Campus Marketplace is running!",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	// Группа маршрутов API v1
	apiV1 := router.Group("/api/v1")
	{
		auth := apiV1.Group("/auth")
		{
			auth.POST("/register", userHandler.Register)
			auth.POST("/login", userHandler.Login) // Добавили эту строку
		}
	}

	// 5. Настройка и запуск HTTP сервера
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Запуск сервера в отдельной горутине, чтобы не блокировать основной поток
	go func() {
		log.Printf("Server started on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// 6. Graceful Shutdown (Корректное завершение работы)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // Блокируемся до получения сигнала (Ctrl+C или Docker stop)

	log.Println("Shutting down server...")

	// Даем серверу 5 секунд на завершение текущих запросов
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
