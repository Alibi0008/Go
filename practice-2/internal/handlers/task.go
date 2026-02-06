package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type Task struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

// tasks - хранилище задач
var tasks = []Task{
	{ID: 1, Title: "Learn Go", Done: false},
}

// nextID - глобальный счетчик для генерации уникальных ID
var nextID = 2

// TaskHandler - это "регулировщик". Он смотрит на метод запроса (GET или POST)
// TaskHandler - теперь обрабатывает и PATCH
func TaskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		GetTasks(w, r)
	case "POST":
		CreateTask(w, r)
	case "PATCH":
		UpdateTask(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func GetTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Проверка query параметра ?id=...
	idStr := r.URL.Query().Get("id")
	if idStr != "" {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid id"})
			return
		}

		for _, task := range tasks {
			if task.ID == id {
				json.NewEncoder(w).Encode(task)
				return
			}
		}

		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "task not found"})
		return
	}

	// Если ID нет, возвращаем все задачи
	json.NewEncoder(w).Encode(tasks)
}

func CreateTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var newTask Task

	// 1. Декодируем JSON из тела запроса в структуру
	// Если пришлют мусор вместо JSON, будет ошибка
	if err := json.NewDecoder(r.Body).Decode(&newTask); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"})
		return
	}

	// 2. Валидация: Title не должен быть пустым
	if newTask.Title == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid title"})
		return
	}

	// 3. Присваиваем ID и дефолтные значения
	newTask.ID = nextID
	nextID++             // Увеличиваем счетчик для следующей задачи
	newTask.Done = false // По заданию всегда false при создании

	// 4. Добавляем в хранилище
	tasks = append(tasks, newTask)

	// 5. Возвращаем ответ 201 Created и созданную задачу
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newTask)
}

func UpdateTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 1. Получаем ID из параметров (например, /tasks?id=1)
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid id"})
		return
	}

	// 2. Читаем JSON с новым статусом (например, {"done": true})
	var updateData struct {
		Done bool `json:"done"`
	}
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"})
		return
	}

	// 3. Ищем задачу и обновляем
	for i, task := range tasks {
		if task.ID == id {
			// Обновляем статус
			tasks[i].Done = updateData.Done

			// Возвращаем ответ об успехе
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]bool{"updated": true})
			return
		}
	}

	// 4. Если не нашли
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]string{"error": "task not found"})
}
