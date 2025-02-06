package repository

import (
	"database/sql"
	"fmt"

	"github.com/shuvo-paul/uptimebot/internal/auth/model"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) SaveUser(user *model.User) (*model.User, error) {
	query := `INSERT INTO user (name, email, password) VALUES (?, ?, ?)`
	result, err := r.db.Exec(query, user.Name, user.Email, user.Password)
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
	query := "SELECT EXISTS(SELECT 1 FROM user WHERE email = ?)"
	err := r.db.QueryRow(query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error checking email existence: %w", err)
	}
	return exists, nil
}

func (r *UserRepository) GetUserByEmail(email string) (*model.User, error) {
	user := &model.User{}
	query := `SELECT id, name, email, password FROM user WHERE email = ?`
	err := r.db.QueryRow(query, email).Scan(&user.ID, &user.Name, &user.Email, &user.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	return user, nil
}

func (r *UserRepository) GetUserByID(id int) (*model.User, error) {
	user := &model.User{}
	query := `SELECT id, name, email from user WHERE id = ?`
	err := r.db.QueryRow(query, id).Scan(&user.ID, &user.Name, &user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	return user, nil
}

type UserRepositoryInterface interface {
	SaveUser(user *model.User) (*model.User, error)
	EmailExists(email string) (bool, error)
	GetUserByEmail(email string) (*model.User, error)
	GetUserByID(id int) (*model.User, error)
}

var _ UserRepositoryInterface = (*UserRepository)(nil)
