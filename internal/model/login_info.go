package model

import "time"

// LoginInfo represents a login record in the database.
type LoginInfo struct {
	InfoID        int        `json:"info_id"`
	UserID        int64      `json:"user_id"`
	Username      string     `json:"username"`
	IP            *string    `json:"ip"`
	LoginLocation *string    `json:"login_location"`
	Browser       *string    `json:"browser"`
	OS            *string    `json:"os"`
	Status        string     `json:"status"`
	Message       *string    `json:"message"`
	LoginAt       *time.Time `json:"login_at"`
}
