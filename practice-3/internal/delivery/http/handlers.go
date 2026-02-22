package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"practice-3/internal/repository/_postgres/users"
	"practice-3/internal/usecase"
	"practice-3/pkg/modules"
	"strconv"
)

type Handler struct {
	usecase *usecase.UserUsecase
}

func NewHandler(u *usecase.UserUsecase) *Handler {
	return &Handler{usecase: u}
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func (h *Handler) writeError(w http.ResponseWriter, status int, err error) {
	h.writeJSON(w, status, map[string]string{"error": err.Error()})
}

// Healthcheck endpoint
func (h *Handler) Healthcheck(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "pass", "description": "User API is healthy"})
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user modules.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		h.writeError(w, http.StatusBadRequest, err)
		return
	}
	id, err := h.usecase.CreateUser(r.Context(), &user)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err)
		return
	}
	h.writeJSON(w, http.StatusCreated, map[string]int{"id": id})
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	user, err := h.usecase.GetUserByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, users.ErrUserNotFound) {
			h.writeError(w, http.StatusNotFound, err)
			return
		}
		h.writeError(w, http.StatusInternalServerError, err)
		return
	}
	h.writeJSON(w, http.StatusOK, user)
}

func (h *Handler) GetUsers(w http.ResponseWriter, r *http.Request) {
	usersList, err := h.usecase.GetUsers(r.Context())
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err)
		return
	}
	h.writeJSON(w, http.StatusOK, usersList)
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	var user modules.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		h.writeError(w, http.StatusBadRequest, err)
		return
	}
	user.ID = id

	if err := h.usecase.UpdateUser(r.Context(), &user); err != nil {
		if errors.Is(err, users.ErrUserNotFound) {
			h.writeError(w, http.StatusNotFound, err)
			return
		}
		h.writeError(w, http.StatusInternalServerError, err)
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]string{"message": "user updated successfully"})
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	if err := h.usecase.DeleteUser(r.Context(), id); err != nil {
		if errors.Is(err, users.ErrUserNotFound) {
			h.writeError(w, http.StatusNotFound, err)
			return
		}
		h.writeError(w, http.StatusInternalServerError, err)
		return
	}
	h.writeJSON(w, http.StatusNoContent, nil)
}
