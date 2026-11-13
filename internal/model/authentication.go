package model // import "admin-console/internal/model"
import "time"

// LoginRequest represents the payload for a login request.
type LoginRequest struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	Uuid      string `json:"uuid"`
	Code      string `json:"code"`
	AutoLogin bool   `json:"auto_login"`
}

// LoginUser represents the user's detail who has logged in.
type LoginUser struct {
	UserID      int64      `json:"user_id"`
	LoginUUID   string     `json:"login_uuid"`
	LastLoginAt *time.Time `json:"last_login_at"`
	ExpiredAt   *time.Time `json:"expired_at"`
	ClientIP    string     `json:"client_ip"`
	Location    string     `json:"location"`
}

func NewLoginUser(user *User, loginUUID string) *LoginUser {
	if user == nil {
		return nil
	}
	lastLoginAt := time.Now().UTC()
	expired_at := lastLoginAt.Add(30 * time.Minute).UTC()
	return &LoginUser{
		UserID:      user.UserID,
		LoginUUID:   loginUUID,
		LastLoginAt: &lastLoginAt,
		ExpiredAt:   &expired_at,
		ClientIP:    "",
		Location:    "",
	}
}
