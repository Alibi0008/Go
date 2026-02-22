package main

import (
	"context"
	"log"
	"net/http"
	"time"

	delivery "practice-3/internal/delivery/http"
	"practice-3/internal/repository/_postgres"
	"practice-3/internal/repository/_postgres/users"
	"practice-3/internal/usecase"
	"practice-3/pkg/modules"
)

func initPostgreConfig() *modules.PostgreConfig {
	return &modules.PostgreConfig{
		Host:        "localhost",
		Port:        "5432",
		Username:    "postgres",
		Password:    "postgres",
		DBName:      "mydb",
		SSLMode:     "disable",
		ExecTimeout: 5 * time.Second,
	}
}

func main() {
	ctx := context.Background()
	dbConfig := initPostgreConfig()

	// Инициализация слоев
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

	// Защита пользовательских роутов ключом API
	mux.Handle("/users/", delivery.AuthMiddleware(apiMux))
	mux.Handle("/users", delivery.AuthMiddleware(apiMux))

	// Оборачиваем все роуты логгером (выполняется первым, оборачивая Auth)
	loggedMux := delivery.LoggingMiddleware(mux)

	log.Println("Сервер запущен на :8080")
	if err := http.ListenAndServe(":8080", loggedMux); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
