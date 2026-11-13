package model // import "admin-console/internal/model"

// LoginRequest represents the payload for a login request.
type LoginRequest struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	Uuid      string `json:"uuid"`
	Code      string `json:"code"`
	AutoLogin bool   `json:"autoLogin"`
}
