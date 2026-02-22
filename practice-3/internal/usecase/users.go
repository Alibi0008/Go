package usecase

import (
	"context"
	"practice-3/internal/repository/_postgres/users"
	"practice-3/pkg/modules"
)

type UserUsecase struct {
	repo *users.Repository
}

func NewUserUsecase(repo *users.Repository) *UserUsecase {
	return &UserUsecase{repo: repo}
}

func (u *UserUsecase) CreateUser(ctx context.Context, user *modules.User) (int, error) {
	return u.repo.CreateUser(ctx, user)
}

func (u *UserUsecase) GetUserByID(ctx context.Context, id int) (*modules.User, error) {
	return u.repo.GetUserByID(ctx, id)
}

func (u *UserUsecase) GetUsers(ctx context.Context) ([]modules.User, error) {
	return u.repo.GetUsers(ctx)
}

func (u *UserUsecase) UpdateUser(ctx context.Context, user *modules.User) error {
	return u.repo.UpdateUser(ctx, user)
}

func (u *UserUsecase) DeleteUser(ctx context.Context, id int) error {
	return u.repo.DeleteUser(ctx, id)
}
