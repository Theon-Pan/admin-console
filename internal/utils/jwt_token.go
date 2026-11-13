package utils // import "admin-console/internal/utils"
import "github.com/golang-jwt/jwt/v5"

const (
	JWTSecret = "asbcdefghijkmnlopqrstuvwxyz"
)

type TokenClaims struct {
	UserID   string `json:"user_id"`
	UserName string `json:"username"`
	UUID     string `json:"uuid"`
	jwt.RegisteredClaims
}

func GenerateToken(claims TokenClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	return token.SignedString([]byte(JWTSecret))
}

func ValidateToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(JWTSecret), nil
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
