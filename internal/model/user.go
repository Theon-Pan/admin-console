package model

import "time"

// User represents a user record in the database
type User struct {
	ID          int        `json:"id"`
	UserID      int64      `json:"user_id"`
	Username    string     `json:"username"`
	Nickname    *string    `json:"nickname"`
	Email       *string    `json:"email"`
	Phonenumber *string    `json:"phonenumber"`
	Password    string     `json:"-"`
	IsAdmin     bool       `json:"is_admin"`
	LastLoginAt *time.Time `json:"last_login_at"`
	Remark      *string    `json:"remark"`
}

// UserCreationRequest represents the request structure for creating a new user
type UserCreationRequest struct {
	UserID      int64  `json:"user_id"`
	Username    string `json:"username"`
	Nickname    string `json:"nickname"`
	Email       string `json:"email"`
	Phonenumber string `json:"phonenumber"`
	Password    string `json:"password"`
	IsAdmin     bool   `json:"is_admin"`
	Remark      string `json:"remark"`
}

// UserModificationRequest represents the request structure for updating a user
type UserModificationRequest struct {
	ID          *int       `json:"id"`
	UserID      *int64     `json:"user_id"`
	Username    *string    `json:"username"`
	Nickname    *string    `json:"nickname"`
	Email       *string    `json:"email"`
	Phonenumber *string    `json:"phonenumber"`
	Password    *string    `json:"password"`
	IsAdmin     *bool      `json:"is_admin"`
	LastLoginAt *time.Time `json:"last_login_at"`
	Remark      *string    `json:"remark"`
}
