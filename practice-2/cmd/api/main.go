package main

import (
	"fmt"
	"net/http"

	"assignment-1/internal/handlers"
	"assignment-1/internal/middleware"
)

func main() {
	http.HandleFunc("/tasks", middleware.Authentication(handlers.TaskHandler))

	fmt.Println("Server is running on http://localhost:8080")

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
