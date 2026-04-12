package repo

import (
	"fmt"
	"strings"
	"sync"

	"practice7/internal/entity"

	"github.com/google/uuid"
)

type UserRepo struct {
	mu            sync.RWMutex
	usersByID     map[uuid.UUID]*entity.User
	idsByUsername map[string]uuid.UUID
	idsByEmail    map[string]uuid.UUID
}

func NewUserRepo() *UserRepo {
	return &UserRepo{
		usersByID:     make(map[uuid.UUID]*entity.User),
		idsByUsername: make(map[string]uuid.UUID),
		idsByEmail:    make(map[string]uuid.UUID),
	}
}

func (u *UserRepo) RegisterUser(user *entity.User) (*entity.User, error) {
	u.mu.Lock()
	defer u.mu.Unlock()

	usernameKey := strings.ToLower(strings.TrimSpace(user.Username))
	emailKey := strings.ToLower(strings.TrimSpace(user.Email))

	if _, exists := u.idsByUsername[usernameKey]; exists {
		return nil, fmt.Errorf("username already exists")
	}

	if _, exists := u.idsByEmail[emailKey]; exists {
		return nil, fmt.Errorf("email already exists")
	}

	userCopy := cloneUser(user)
	u.usersByID[userCopy.ID] = userCopy
	u.idsByUsername[usernameKey] = userCopy.ID
	u.idsByEmail[emailKey] = userCopy.ID

	return cloneUser(userCopy), nil
}

func (u *UserRepo) GetByUsername(username string) (*entity.User, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()

	id, exists := u.idsByUsername[strings.ToLower(strings.TrimSpace(username))]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	user, ok := u.usersByID[id]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}

	return cloneUser(user), nil
}

func (u *UserRepo) GetByID(id uuid.UUID) (*entity.User, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()

	user, exists := u.usersByID[id]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	return cloneUser(user), nil
}

func (u *UserRepo) PromoteUser(id uuid.UUID) (*entity.User, error) {
	u.mu.Lock()
	defer u.mu.Unlock()

	user, exists := u.usersByID[id]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	user.Role = entity.RoleAdmin
	return cloneUser(user), nil
}

func cloneUser(user *entity.User) *entity.User {
	if user == nil {
		return nil
	}

	userCopy := *user
	return &userCopy
}
