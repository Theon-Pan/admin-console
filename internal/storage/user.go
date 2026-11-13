package storage // import "admin-console/internal/storage"
import (
	"admin-console/internal/crypto"
	"admin-console/internal/model"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var dummyBcryptHash = []byte("$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy")

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

// UserByID finds a user by the ID.
func (s *Storage) UserByID(userID int64) (*model.User, error) {
	query := `
		SELECT
			id,
			user_id,
			username,
			nickname,
			email,
			phonenumber,
			password,
			is_admin,
			last_login_at,
			remark
		FROM
			users
		where
			user_id = $1
	`

	return s.fetchUser(query, userID)
}

// UserByUsername finds a user by the username.
func (s *Storage) UserByUsername(username string) (*model.User, error) {
	query := `
		SELECT
			id,
			user_id,
			username,
			nickname,
			email,
			phonenumber,
			password,
			is_admin,
			last_login_at,
			remark
		FROM
			users
		where
			username = LOWER($1)
	`

	return s.fetchUser(query, username)
}

// CheckPassword returns nil if the given password matches the user's stored hash.
func (s *Storage) CheckPassword(username, password string) error {
	var hash string
	username = strings.ToLower(username)

	err := s.db.QueryRow("SELECT password FROM users WHERE username=$1", username).Scan(&hash)
	if errors.Is(err, sql.ErrNoRows) {
		// Perform a dummy bcrypt comparison against the hashed `password` string to avoid leaking whether the user exists via response timing.
		_ = bcrypt.CompareHashAndPassword(dummyBcryptHash, []byte(password))
		return fmt.Errorf(`store: unable to find this user: %s`, username)
	} else if err != nil {
		return fmt.Errorf(`store: unable to fetch user: %v`, err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return fmt.Errorf(`store: invalid password for "%s" (%v)`, username, err)
	}

	return nil
}

func (s *Storage) fetchUser(query string, args ...any) (*model.User, error) {
	var user model.User
	err := s.db.QueryRow(query, args...).Scan(
		&user.ID,
		&user.UserID,
		&user.Username,
		&user.Nickname,
		&user.Email,
		&user.Phonenumber,
		&user.Password,
		&user.IsAdmin,
		&user.LastLoginAt,
		&user.Remark,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf(`store: unable to fetch user: %v`, err)
	}

	return &user, nil
}
