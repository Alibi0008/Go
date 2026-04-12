package usecase

import "practice7/internal/entity"

type UserInterface interface {
	RegisterUser(input *entity.CreateUserDTO) (*entity.UserResponse, string, error)
	LoginUser(input *entity.LoginUserDTO) (string, error)
	GetMe(userID string) (*entity.UserResponse, error)
	PromoteUser(userID string) (*entity.UserResponse, error)
}
