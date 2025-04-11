package repository

import (
	"fmt"

	"github.com/shuvo-paul/uptimebot/internal/auth/model"
	"github.com/shuvo-paul/uptimebot/internal/database"
)

type UserRepository struct {
	db database.Querier
}

func NewUserRepository(db database.Querier) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) SaveUser(user *model.User) (*model.User, error) {
	query := `INSERT INTO usr (email, password) VALUES ($1, $2) RETURNING id`
	err := r.db.QueryRow(query, user.Email, user.Password).Scan(&user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	return user, nil
}

func (r *UserRepository) EmailExists(email string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM usr WHERE email = $1)"
	err := r.db.QueryRow(query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error checking email existence: %w", err)
	}
	return exists, nil
}

func (r *UserRepository) GetUserByEmail(email string) (*model.User, error) {
	user := &model.User{}
	query := `SELECT id, email, password FROM usr WHERE email = $1`
	err := r.db.QueryRow(query, email).Scan(&user.ID, &user.Email, &user.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	return user, nil
}

func (r *UserRepository) GetUserByID(id int) (*model.User, error) {
	user := &model.User{}
	query := `SELECT id, email, verified from usr WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(&user.ID, &user.Email, &user.Verified)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	return user, nil
}

func (r *UserRepository) UpdateUser(user *model.User) (*model.User, error) {
	query := `UPDATE usr SET email = $1, verified = $2 WHERE id = $3`
	result, err := r.db.Exec(query, user.Email, user.Verified, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return nil, fmt.Errorf("no user found with ID: %d", user.ID)
	}

	return user, nil
}

func (r *UserRepository) UpdatePassword(userID int, hashedPassword string) error {
	query := `UPDATE usr SET password = $1 WHERE id = $2`
	result, err := r.db.Exec(query, hashedPassword, userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no user found with ID: %d", userID)
	}

	return nil
}

type UserRepositoryInterface interface {
	SaveUser(user *model.User) (*model.User, error)
	EmailExists(email string) (bool, error)
	GetUserByEmail(email string) (*model.User, error)
	GetUserByID(id int) (*model.User, error)
	UpdateUser(user *model.User) (*model.User, error)
	UpdatePassword(userID int, hashedPassword string) error
}

var _ UserRepositoryInterface = (*UserRepository)(nil)
