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

func (s *Storage) StoreLoginUser(ctx context.Context, loginUser model.LoginUser) error {
	data, err := json.Marshal(loginUser)
	if err != nil {
		return err
	}

	return s.redisClient.Set(ctx, "login_tokens:"+loginUser.LoginUUID, data, 30*time.Minute).Err()
}

func (s *Storage) RemoveLoginUser(ctx context.Context, loginUser model.LoginUser) error {
	return s.redisClient.Del(ctx, "login_tokens:"+loginUser.LoginUUID).Err()
}

func (s *Storage) RetrieveLoginUser(ctx context.Context, id string) (*model.LoginUser, error) {
	return s.getLoginUser(ctx, id)
}

func (s *Storage) getCaptchaCode(ctx context.Context, id string) (string, error) {
	return s.redisClient.Get(ctx, "captcha_codes:"+id).Result()
}

func (s *Storage) getLoginUser(ctx context.Context, id string) (*model.LoginUser, error) {
	data, err := s.redisClient.Get(ctx, "login_tokens:"+id).Result()
	if err != nil {
		return nil, err
	}
	var loginUser model.LoginUser
	if err := json.Unmarshal([]byte(data), &loginUser); err != nil {
		return nil, err
	}
	return &loginUser, nil
}
