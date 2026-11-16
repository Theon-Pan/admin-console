package utils // import "admin-console/internal/utils"
import (
	"admin-console/internal/config"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

const (
	JWTSecretLength = 64
	JWTSecret       = "asbcdefghijkmnlopqrstuvwxyz"
)

type TokenClaims struct {
	UserID   string `json:"user_id"`
	UserName string `json:"username"`
	UUID     string `json:"uuid"`
	jwt.RegisteredClaims
}

func GenerateJWTSecret() ([]byte, error) {
	secret := make([]byte, JWTSecretLength)
	if _, err := rand.Read(secret); err != nil {
		return nil, fmt.Errorf("generate JWT secret: %v", err)
	}

	return secret, nil
}

func LoadJWTSecretFromFile(path string) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}

	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("JWT secret file must be a regular file, not a symlink or directory")
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	opened, err := file.Stat()
	if err != nil {
		return nil, err
	}

	if !os.SameFile(info, opened) {
		return nil, fmt.Errorf("JWT secret file changed while opening")
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	// if len(data) != JWTSecretLength {
	// 	return nil, fmt.Errorf("Secret file contain exactly %d raw randombytes", JWTSecretLength)
	// }

	return data, nil
}

func GenerateToken(claims TokenClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	jwtSecret, err := getJWTSecret()
	if err != nil {
		return "", err
	}

	return token.SignedString(jwtSecret)
}

func ValidateToken(tokenString string) (*TokenClaims, error) {
	jwtSecret, err := getJWTSecret()
	if err != nil {
		return nil, err
	}
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}

func TokenHeader(tokenstring string) (map[string]any, error) {
	token, _, err := jwt.NewParser().ParseUnverified(tokenstring, jwt.MapClaims{})
	if err != nil {
		return nil, err
	}

	return token.Header, nil
}

func getJWTSecret() ([]byte, error) {
	jwtSecret := config.Opts.JWTSecret()
	if strings.TrimSpace(jwtSecret) == "" {
		return nil, fmt.Errorf("JWT secret is not set")
	}
	return base64.StdEncoding.DecodeString(jwtSecret)
}
