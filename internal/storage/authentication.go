package storage // import "admin-console/internal/storage"
import (
	"context"
	"time"
)

func (s *Storage) StoreCaptchaCode(context context.Context, id, code string) error {
	return s.redisClient.Set(context, "captcha_codes:"+id, code, 30*time.Minute).Err()
}

func (s *Storage) GetCaptchaCode(context context.Context, id string) (string, error) {
	return s.redisClient.Get(context, "captcha_codes:"+id).Result()
}
