package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"go-users-project/repository"
)

type Handler struct {
	repo *repository.Repository
}

func NewHandler(repo *repository.Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	page, _ := strconv.Atoi(query.Get("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(query.Get("pageSize"))
	if pageSize < 1 {
		pageSize = 10
	}

	orderBy := query.Get("order_by")

	filters := make(map[string]string)
	for key, values := range query {
		if key != "page" && key != "pageSize" && key != "order_by" && len(values) > 0 && values[0] != "" {
			filters[key] = values[0]
		}
	}

	response, err := h.repo.GetPaginatedUsers(page, pageSize, filters, orderBy)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) GetCommonFriendsHandler(w http.ResponseWriter, r *http.Request) {
	user1Str := r.URL.Query().Get("user1")
	user2Str := r.URL.Query().Get("user2")

	if user1Str == "" || user2Str == "" {
		http.Error(w, "Параметры user1 и user2 обязательны", http.StatusBadRequest)
		return
	}

	user1, err1 := strconv.Atoi(user1Str)
	user2, err2 := strconv.Atoi(user2Str)

	if err1 != nil || err2 != nil {
		http.Error(w, "ID пользователей должны быть числами", http.StatusBadRequest)
		return
	}

	friends, err := h.repo.GetCommonFriends(user1, user2)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(friends)
}
