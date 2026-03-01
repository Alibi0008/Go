package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	delivery "practice-3/internal/delivery/http"
	"practice-3/internal/repository/_postgres"
	"practice-3/internal/repository/_postgres/users"
	"practice-3/internal/usecase"
	"practice-3/pkg/modules"
)

// initPostgreConfig теперь считывает настройки из Docker Compose (Practice 4)
func initPostgreConfig() *modules.PostgreConfig {
	return &modules.PostgreConfig{
		// Используем имя сервиса "db" из docker-compose вместо "localhost"
		Host:        getEnv("DB_HOST", "db"),
		Port:        getEnv("DB_PORT", "5432"),
		Username:    getEnv("DB_USER", "postgres"),
		Password:    getEnv("DB_PASSWORD", "postgres"),
		DBName:      getEnv("DB_NAME", "mydb"),
		SSLMode:     getEnv("DB_SSLMODE", "disable"),
		ExecTimeout: 5 * time.Second,
	}
}

// Вспомогательная функция для получения переменных окружения с дефолтным значением
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {
	ctx := context.Background()
	dbConfig := initPostgreConfig()

	// Инициализация слоев (Dependency Injection)
	pgxDialect := _postgres.NewPGXDialect(ctx, dbConfig)
	userRepo := users.NewUserRepository(pgxDialect)
	userUsecase := usecase.NewUserUsecase(userRepo)
	handler := delivery.NewHandler(userUsecase)

	// Настройка маршрутизатора (Go 1.22+)
	mux := http.NewServeMux()

	// Открытый эндпоинт Healthcheck
	mux.HandleFunc("GET /health", handler.Healthcheck)

	// Маршруты для пользователей
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("POST /users", handler.CreateUser)
	apiMux.HandleFunc("GET /users", handler.GetUsers)
	apiMux.HandleFunc("GET /users/{id}", handler.GetUser)
	apiMux.HandleFunc("PUT /users/{id}", handler.UpdateUser)
	apiMux.HandleFunc("DELETE /users/{id}", handler.DeleteUser)

	// Защита пользовательских роутов через Middleware (Auth)
	mux.Handle("/users/", delivery.AuthMiddleware(apiMux))
	mux.Handle("/users", delivery.AuthMiddleware(apiMux))

	// Оборачиваем все роуты логгером (Middleware: Logging)
	loggedMux := delivery.LoggingMiddleware(mux)

	// Согласно сценарию демо-видео, выводим фразу "Starting the Server"
	log.Println("Starting the Server on :8080...")
	if err := http.ListenAndServe(":8080", loggedMux); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
