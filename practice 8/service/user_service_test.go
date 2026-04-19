package service

import (
	"errors"
	"testing"

	"practice-8/repository"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestGetUserByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockUserRepository(ctrl)
	userService := NewUserService(mockRepo)
	user := &repository.User{ID: 1, Name: "Bakytzhan Agai", Email: "bakytzhan@example.com"}

	mockRepo.EXPECT().GetUserByID(1).Return(user, nil)

	result, err := userService.GetUserByID(1)

	assert.NoError(t, err)
	assert.Equal(t, user, result)
}

func TestCreateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockUserRepository(ctrl)
	userService := NewUserService(mockRepo)
	user := &repository.User{ID: 1, Name: "Bakytzhan Agai", Email: "bakytzhan@example.com"}

	mockRepo.EXPECT().CreateUser(user).Return(nil)

	err := userService.CreateUser(user)

	assert.NoError(t, err)
}

func TestRegisterUser_UserAlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockUserRepository(ctrl)
	userService := NewUserService(mockRepo)
	user := &repository.User{ID: 2, Name: "Aruzhan", Email: "aruzhan@example.com"}
	existing := &repository.User{ID: 7, Name: "Existing", Email: "aruzhan@example.com"}

	mockRepo.EXPECT().GetByEmail("aruzhan@example.com").Return(existing, nil)

	err := userService.RegisterUser(user, "aruzhan@example.com")

	assert.EqualError(t, err, "user with this email already exists")
}

func TestRegisterUser_NewUserSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockUserRepository(ctrl)
	userService := NewUserService(mockRepo)
	user := &repository.User{ID: 2, Name: "Aruzhan", Email: "aruzhan@example.com"}

	gomock.InOrder(
		mockRepo.EXPECT().GetByEmail("aruzhan@example.com").Return(nil, nil),
		mockRepo.EXPECT().CreateUser(user).Return(nil),
	)

	err := userService.RegisterUser(user, "aruzhan@example.com")

	assert.NoError(t, err)
}

func TestRegisterUser_CreateFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockUserRepository(ctrl)
	userService := NewUserService(mockRepo)
	user := &repository.User{ID: 2, Name: "Aruzhan", Email: "aruzhan@example.com"}

	gomock.InOrder(
		mockRepo.EXPECT().GetByEmail("aruzhan@example.com").Return(nil, nil),
		mockRepo.EXPECT().CreateUser(user).Return(errors.New("create failed")),
	)

	err := userService.RegisterUser(user, "aruzhan@example.com")

	assert.EqualError(t, err, "create failed")
}

func TestUpdateUserName_EmptyName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockUserRepository(ctrl)
	userService := NewUserService(mockRepo)

	err := userService.UpdateUserName(2, "")

	assert.EqualError(t, err, "name cannot be empty")
}

func TestUpdateUserName_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockUserRepository(ctrl)
	userService := NewUserService(mockRepo)

	mockRepo.EXPECT().GetUserByID(99).Return(nil, errors.New("user not found"))

	err := userService.UpdateUserName(99, "Dana")

	assert.EqualError(t, err, "user not found")
}

func TestUpdateUserName_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockUserRepository(ctrl)
	userService := NewUserService(mockRepo)
	user := &repository.User{ID: 4, Name: "Old Name", Email: "user@example.com"}

	mockRepo.EXPECT().GetUserByID(4).Return(user, nil)
	mockRepo.EXPECT().UpdateUser(gomock.Any()).DoAndReturn(func(updated *repository.User) error {
		assert.Equal(t, 4, updated.ID)
		assert.Equal(t, "New Name", updated.Name)
		assert.Equal(t, "user@example.com", updated.Email)
		return nil
	})

	err := userService.UpdateUserName(4, "New Name")

	assert.NoError(t, err)
	assert.Equal(t, "New Name", user.Name)
}

func TestUpdateUserName_UpdateFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockUserRepository(ctrl)
	userService := NewUserService(mockRepo)
	user := &repository.User{ID: 4, Name: "Old Name", Email: "user@example.com"}

	mockRepo.EXPECT().GetUserByID(4).Return(user, nil)
	mockRepo.EXPECT().UpdateUser(gomock.Any()).DoAndReturn(func(updated *repository.User) error {
		assert.Equal(t, "Updated Name", updated.Name)
		return errors.New("update failed")
	})

	err := userService.UpdateUserName(4, "Updated Name")

	assert.EqualError(t, err, "update failed")
	assert.Equal(t, "Updated Name", user.Name)
}

func TestDeleteUser_AdminForbidden(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockUserRepository(ctrl)
	userService := NewUserService(mockRepo)

	err := userService.DeleteUser(1)

	assert.EqualError(t, err, "it is not allowed to delete admin user")
}

func TestDeleteUser_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockUserRepository(ctrl)
	userService := NewUserService(mockRepo)

	deletedID := 0
	mockRepo.EXPECT().DeleteUser(2).DoAndReturn(func(id int) error {
		deletedID = id
		return nil
	})

	err := userService.DeleteUser(2)

	assert.NoError(t, err)
	assert.Equal(t, 2, deletedID)
}

func TestDeleteUser_RepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockUserRepository(ctrl)
	userService := NewUserService(mockRepo)

	mockRepo.EXPECT().DeleteUser(2).Return(errors.New("delete failed"))

	err := userService.DeleteUser(2)

	assert.EqualError(t, err, "delete failed")
}
