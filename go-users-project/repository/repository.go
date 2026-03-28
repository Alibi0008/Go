package repository

import (
	"database/sql"
	"fmt"
	"go-users-project/models"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetPaginatedUsers(page, pageSize int, filters map[string]string, orderBy string) (models.PaginatedResponse, error) {
	allowedColumns := map[string]bool{
		"id": true, "name": true, "email": true, "gender": true, "birth_date": true,
	}

	baseQuery := "FROM users WHERE 1=1"
	var args []interface{}
	argID := 1

	for key, val := range filters {
		if allowedColumns[key] {
			baseQuery += fmt.Sprintf(" AND %s = $%d", key, argID)
			args = append(args, val)
			argID++
		}
	}

	var totalCount int
	err := r.db.QueryRow("SELECT COUNT(*) "+baseQuery, args...).Scan(&totalCount)
	if err != nil {
		return models.PaginatedResponse{}, err
	}

	if orderBy == "" || !allowedColumns[orderBy] {
		orderBy = "id"
	}

	offset := (page - 1) * pageSize
	dataQuery := fmt.Sprintf("SELECT id, name, email, gender, birth_date %s ORDER BY %s LIMIT $%d OFFSET $%d", baseQuery, orderBy, argID, argID+1)
	args = append(args, pageSize, offset)

	rows, err := r.db.Query(dataQuery, args...)
	if err != nil {
		return models.PaginatedResponse{}, err
	}
	defer rows.Close()

	users := []models.User{}
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Gender, &u.BirthDate); err != nil {
			return models.PaginatedResponse{}, err
		}
		users = append(users, u)
	}

	return models.PaginatedResponse{
		Data:       users,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

func (r *Repository) GetCommonFriends(userID1, userID2 int) ([]models.User, error) {
	query := `
		SELECT u.id, u.name, u.email, u.gender, u.birth_date
		FROM users u
		JOIN user_friends uf1 ON u.id = uf1.friend_id
		JOIN user_friends uf2 ON u.id = uf2.friend_id
		WHERE uf1.user_id = $1 AND uf2.user_id = $2
	`
	rows, err := r.db.Query(query, userID1, userID2)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	friends := []models.User{}
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Gender, &u.BirthDate); err != nil {
			return nil, err
		}
		friends = append(friends, u)
	}
	return friends, nil
}
