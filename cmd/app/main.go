package main

import (
	"Avito-back/internal/delivery/http/middleware"
	"Avito-back/internal/repository/redis"
	"Avito-back/internal/repository/s3"
	"Avito-back/internal/usecase/ad"
	"Avito-back/internal/usecase/chat"
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

	s3Endpoint := os.Getenv("MINIO_ENDPOINT")       // "minio:9000"
	s3AccessKey := os.Getenv("MINIO_ROOT_USER")     // "minioadmin"
	s3SecretKey := os.Getenv("MINIO_ROOT_PASSWORD") // "minioadmin"
	s3Bucket := "campus-images"                     // Название папки для фото

	// 2. Инициализируем S3 Репозиторий
	s3Repo, err := s3.NewFileRepository(s3Endpoint, s3AccessKey, s3SecretKey, s3Bucket)
	if err != nil {
		log.Fatalf("Failed to connect to MinIO: %v", err)
	}

	// Repository (Слой работы с БД)
	userRepo := postgres.NewUserRepository(db)
	adRepo := postgres.NewAdRepository(db)

	// Usecase (Слой бизнес-логики)
	userUsecase := user.NewUserUsecase(userRepo)

	cacheRepo := redis.NewAdCacheRepository(os.Getenv("REDIS_HOST")) // "redis:6379"
	adUsecase := ad.NewAdUsecase(adRepo, s3Repo, cacheRepo)

	// Delivery/Handler (Слой HTTP интерфейса)
	userHandler := &v1.UserHandler{
		Usecase: userUsecase,
	}
	adHandler := &v1.AdHandler{Usecase: adUsecase}

	// Инициализация
	chatRepo := postgres.NewChatRepository(db)
	chatUsecase := chat.NewChatUsecase(chatRepo, adRepo)
	chatHandler := &v1.ChatHandler{Usecase: chatUsecase}

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

	jwtSecret := "your-super-secret-key-for-campus"

	// Группа маршрутов API v1
	apiV1 := router.Group("/api/v1")
	{
		auth := apiV1.Group("/auth")
		{
			auth.POST("/register", userHandler.Register)
			auth.POST("/login", userHandler.Login) // Добавили эту строку
		}

		// ЗАЩИЩЕННЫЕ МАРШРУТЫ (только для авторизованных)
		protected := apiV1.Group("/")
		protected.Use(middleware.AuthMiddleware(jwtSecret))
		{
			protected.POST("/ads", adHandler.Create) // Создать объявление
			protected.POST("/ads/:id/images", adHandler.UploadImage)
			// Роуты
			protected.POST("/chats/messages", chatHandler.Send)           // Отправить сообщение
			protected.GET("/chats/:id/messages", chatHandler.GetMessages) // Посмотреть переписку
			protected.GET("/chats", chatHandler.GetMyChats)
		}

		// ОТКРЫТЫЕ МАРШРУТЫ (смотреть могут все)
		apiV1.GET("/ads/:id", adHandler.GetByID)
		apiV1.GET("/ads", adHandler.List)
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
