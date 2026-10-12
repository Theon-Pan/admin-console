package crypto // import "admin-console/internal/crypto"
import (
	"encoding/hex"
	"hash/fnv"

	"golang.org/x/crypto/bcrypt"
)

// HashFromBytes returns a non-cryptographic checksum of the input
func HashFromBytes(value []byte) string {
	h := fnv.New128a()
	h.Write(value)
	return hex.EncodeToString(h.Sum(nil))
}
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}
