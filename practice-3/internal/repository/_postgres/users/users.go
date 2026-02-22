package users

import (
	"context"
	"database/sql"
	"errors"
	"practice-3/internal/repository/_postgres"
	"practice-3/pkg/modules"
)

var ErrUserNotFound = errors.New("user not found")

type Repository struct {
	db *_postgres.Dialect
}

func NewUserRepository(db *_postgres.Dialect) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateUser(ctx context.Context, user *modules.User) (int, error) {
	query := `INSERT INTO users (name, email, age) VALUES ($1, $2, $3) RETURNING id`
	var id int
	err := r.db.DB.QueryRowContext(ctx, query, user.Name, user.Email, user.Age).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *Repository) GetUserByID(ctx context.Context, id int) (*modules.User, error) {
	var user modules.User
	err := r.db.DB.GetContext(ctx, &user, "SELECT id, name, email, age, created_at FROM users WHERE id = $1", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *Repository) GetUsers(ctx context.Context) ([]modules.User, error) {
	var users []modules.User

	err := r.db.DB.SelectContext(ctx, &users, "SELECT id, name, email, age, created_at FROM users")
	if err != nil {
		return nil, err
	}

	if users == nil {
		users = []modules.User{}
	}

	return users, nil
}

func (r *Repository) UpdateUser(ctx context.Context, user *modules.User) error {
	query := `UPDATE users SET name = $1, email = $2, age = $3 WHERE id = $4`
	result, err := r.db.DB.ExecContext(ctx, query, user.Name, user.Email, user.Age, user.ID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *Repository) DeleteUser(ctx context.Context, id int) error {
	query := `DELETE FROM users WHERE id = $1`
	result, err := r.db.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}
