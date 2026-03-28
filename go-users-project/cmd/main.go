package main

import (
	"fmt"
	"log"
	"net/http"

	"go-users-project/db"
	"go-users-project/handlers"
	"go-users-project/repository"
)

func main() {
	// Подключаемся к БД
	database := db.Connect()
	defer database.Close()

	// Инициализируем репозиторий и хендлер
	repo := repository.NewRepository(database)
	h := handlers.NewHandler(repo)

	// Назначаем маршруты
	http.HandleFunc("/users", h.GetUsersHandler)
	http.HandleFunc("/users/common-friends", h.GetCommonFriendsHandler)

	// Запускаем сервер
	fmt.Println("Сервер успешно запущен на порту: 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
