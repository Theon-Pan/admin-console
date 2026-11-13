package storage // import "admin-console/internal/storage"
import (
	"admin-console/internal/model"
	"context"
	"encoding/json"
	"strings"
	"time"
)

func (s *Storage) StoreCaptchaCode(context context.Context, id, code string) error {
	return s.redisClient.Set(context, "captcha_codes:"+id, code, 30*time.Minute).Err()
}

func (s *Storage) VerifyCaptchaCode(context context.Context, id, actualCode string) (bool, error) {
	code, err := s.getCaptchaCode(context, id)
	if err != nil {
		return false, err
	}

	if strings.Compare(code, actualCode) != 0 {
		return false, nil
	}

	return true, nil
}

func (s *Storage) StoreLoginUser(context context.Context, loginUser model.LoginUser) error {
	data, err := json.Marshal(loginUser)
	if err != nil {
		return err
	}

	return s.redisClient.Set(context, "login_tokens:"+loginUser.LoginUUID, data, 30*time.Minute).Err()
}

func (s *Storage) RetrieveLoginUser(context context.Context, id string) (*model.LoginUser, error) {
	return s.getLoginUser(context, id)
}

func (s *Storage) getCaptchaCode(context context.Context, id string) (string, error) {
	return s.redisClient.Get(context, "captcha_codes:"+id).Result()
}

func (s *Storage) getLoginUser(context context.Context, id string) (*model.LoginUser, error) {
	data, err := s.redisClient.Get(context, "login_tokens:"+id).Result()
	if err != nil {
		return nil, err
	}
	var loginUser model.LoginUser
	if err := json.Unmarshal([]byte(data), &loginUser); err != nil {
		return nil, err
	}
	return &loginUser, nil
}
