package usecase

import (
	"fmt"
	"strings"
	"time"

	"practice7/internal/entity"
	"practice7/internal/usecase/repo"
	"practice7/internal/utils"

	"github.com/google/uuid"
)

type UserUseCase struct {
	repo       *repo.UserRepo
	jwtManager *utils.JWTManager
}

func NewUserUseCase(r *repo.UserRepo, jwtManager *utils.JWTManager) *UserUseCase {
	return &UserUseCase{
		repo:       r,
		jwtManager: jwtManager,
	}
}

func (u *UserUseCase) RegisterUser(input *entity.CreateUserDTO) (*entity.UserResponse, string, error) {
	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		return nil, "", fmt.Errorf("hash password: %w", err)
	}

	user := &entity.User{
		ID:        uuid.New(),
		Username:  strings.TrimSpace(input.Username),
		Email:     strings.TrimSpace(strings.ToLower(input.Email)),
		Password:  hashedPassword,
		Role:      entity.RoleUser,
		Verified:  true,
		CreatedAt: time.Now().UTC(),
	}

	createdUser, err := u.repo.RegisterUser(user)
	if err != nil {
		return nil, "", fmt.Errorf("register user: %w", err)
	}

	return entity.NewUserResponse(createdUser), uuid.NewString(), nil
}

func (u *UserUseCase) LoginUser(input *entity.LoginUserDTO) (string, error) {
	userFromRepo, err := u.repo.GetByUsername(input.Username)
	if err != nil {
		return "", fmt.Errorf("get user by username: %w", err)
	}

	if !utils.CheckPassword(userFromRepo.Password, input.Password) {
		return "", fmt.Errorf("invalid username or password")
	}

	token, err := u.jwtManager.GenerateToken(userFromRepo.ID, userFromRepo.Role)
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}

	return token, nil
}

func (u *UserUseCase) GetMe(userID string) (*entity.UserResponse, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("parse user id: %w", err)
	}

	user, err := u.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return entity.NewUserResponse(user), nil
}

func (u *UserUseCase) PromoteUser(userID string) (*entity.UserResponse, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("parse user id: %w", err)
	}

	promotedUser, err := u.repo.PromoteUser(id)
	if err != nil {
		return nil, fmt.Errorf("promote user: %w", err)
	}

	return entity.NewUserResponse(promotedUser), nil
}

func (u *UserUseCase) EnsureBootstrapAdmin(username, email, password string) (bool, *entity.UserResponse, error) {
	existingUser, err := u.repo.GetByUsername(username)
	if err == nil {
		return false, entity.NewUserResponse(existingUser), nil
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return false, nil, fmt.Errorf("hash admin password: %w", err)
	}

	admin := &entity.User{
		ID:        uuid.New(),
		Username:  strings.TrimSpace(username),
		Email:     strings.TrimSpace(strings.ToLower(email)),
		Password:  hashedPassword,
		Role:      entity.RoleAdmin,
		Verified:  true,
		CreatedAt: time.Now().UTC(),
	}

	createdAdmin, createErr := u.repo.RegisterUser(admin)
	if createErr != nil {
		return false, nil, fmt.Errorf("register bootstrap admin: %w", createErr)
	}

	return true, entity.NewUserResponse(createdAdmin), nil
}
