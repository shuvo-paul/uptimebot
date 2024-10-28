package repository

import (
	"database/sql"
	"fmt"

	"github.com/shuvo-paul/sitemonitor/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) SaveUser(user *models.User) (*models.User, error) {
	query := `INSERT INTO users (username, email, password) VALUES (?, ?, ?)`
	result, err := r.db.Exec(query, user.Username, user.Email, user.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	user.ID = int(id)
	return user, nil
}

func (r *UserRepository) EmailExists(email string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)"
	err := r.db.QueryRow(query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error checking email existence: %w", err)
	}
	return exists, nil
}

func (r *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, username, email, password FROM users WHERE email = ?`
	err := r.db.QueryRow(query, email).Scan(&user.ID, &user.Username, &user.Email, &user.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	return user, nil
}

type UserRepositoryInterface interface {
	SaveUser(user *models.User) (*models.User, error)
	EmailExists(email string) (bool, error)
	GetUserByEmail(email string) (*models.User, error)
}

var _ UserRepositoryInterface = (*UserRepository)(nil)
