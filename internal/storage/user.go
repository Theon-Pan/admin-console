package storage // import "admin-console/internal/storage"
import (
	"admin-console/internal/crypto"
	"admin-console/internal/model"
	"fmt"
)

// UserExists returns true if a user with the given username exists.
func (s *Storage) UserExists(username string) bool {
	var result bool
	s.db.QueryRow(`SELECT true FROM users WHERE username=LOWER($1) LIMIT 1`, username).Scan(&result)
	return result
}

// CreateUser creates a new user.
func (s *Storage) CreateUser(userCreationRequest *model.UserCreationRequest) (*model.User, error) {
	var hashedPassword string
	if userCreationRequest.Password != "" {
		var err error
		hashedPassword, err = crypto.HashPassword(userCreationRequest.Password)
		if err != nil {
			return nil, err
		}
	}

	query := `
		INSERT INTO users
			(user_id, username, password, is_admin)
		VALUES
			($1, LOWER($2), $3, $4)
		RETURNING
			id,
			user_id,
			username,
			nickname,
			email,
			phonenumber,
			is_admin,
			last_login_at,
			remark
	`
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf(`store: unable to start transation: %v`, err)
	}

	var user model.User
	err = tx.QueryRow(
		query,
		userCreationRequest.UserID,
		userCreationRequest.Username,
		hashedPassword,
		userCreationRequest.IsAdmin,
	).Scan(
		&user.ID,
		&user.UserID,
		&user.Username,
		&user.Nickname,
		&user.Email,
		&user.Phonenumber,
		&user.IsAdmin,
		&user.LastLoginAt,
		&user.Remark,
	)

	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf(`store: unable to create user: %v`, err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf(`store: unable to commit transation: %v`, err)
	}

	return &user, nil
}
