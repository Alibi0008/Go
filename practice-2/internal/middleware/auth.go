package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Authentication middleware
func Authentication(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// --- 1. ЛОГИРОВАНИЕ (Требование задания) ---
		// Выводим: Время | Метод | Путь
		fmt.Printf("%s %s %s\n", time.Now().Format("2006-01-02 15:04:05"), r.Method, r.URL.Path)

		// --- 2. АВТОРИЗАЦИЯ (Требование задания) ---
		// Читаем заголовок X-API-KEY
		apiKey := r.Header.Get("X-API-KEY")

		// Проверяем ключ (по заданию: "secret12345")
		if apiKey != "secret12345" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized) // 401 ошибка
			json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
			return // Прерываем выполнение, дальше не пускаем
		}

		// Если всё ок, передаем управление дальше (в handlers.TaskHandler)
		next(w, r)
	}
}
