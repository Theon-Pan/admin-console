package model

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
)

// WebSession represents a web session record in the database
type WebSession struct {
	ID         string
	SecretHash []byte
	CreatedAt  time.Time
	UserAgent  string
	IP         string
	userID     *int64
	state      webSessionState
	dirty      bool
}

// WebSessionState stores transient browser session state as a JSON blob.
type webSessionState struct {
	CSRF               string                `json:"csrf,omitempty"`
	SuccessMessage     string                `json:"success_message,omitempty"`
	ErrorMessage       string                `json:"error_message,omitempty"`
	OAuth2             *WebSessionOAuth2     `json:"oauth2,omitempty"`
	WebAuthn           *webauthn.SessionData `json:"webauthn,omitempty"`
	LastForceRefreshAt *time.Time            `json:"last_force_refresh_at,omitempty"`
	Language           string                `json:"language,omitempty"`
	Theme              string                `json:"theme,omitempty"`
}

// WebSessionOAuth2 stores transient OAuth2 flow state.
type WebSessionOAuth2 struct {
	State        string `json:"state,omitempty"`
	CodeVerifier string `json:"code_verifier,omitempty"`
}

// NewWebSession builds an unauthenticated browser session with a fresh
// identify and returns it along with the raw session secret.
func NewWebSession(userAgent, ip string) (*WebSession, string) {
	secret := rand.Text()
	session := &WebSession{
		ID:         rand.Text(),
		SecretHash: hashWebSessionSecret(secret),
		UserAgent:  userAgent,
		IP:         ip,
	}
	session.state.CSRF = rand.Text()

	return session, secret
}

func hashWebSessionSecret(secret string) []byte {
	sum := sha256.Sum256([]byte(secret))
	return sum[:]
}

// VerifySecret reports whether the given raw secret matches the stored hash.
func (s *WebSession) VerifySecret(secret string) bool {
	if secret == "" || len(secret) == 0 {
		return false
	}

	actual := hashWebSessionSecret(secret)
	return subtle.ConstantTimeCompare(actual, s.SecretHash) == 1
}

// MarshalState serializes the session state to JSON for storage.
func (s *WebSession) MarshalState() ([]byte, error) {
	return json.Marshal(s.state)
}

// UnmarshalState populates the session state from raw JSON bytes.
func (s *WebSession) UnmarshalState(data []byte) error {
	s.state = webSessionState{}
	if len(data) == 0 {
		return nil
	}

	return json.Unmarshal(data, &s.state)
}
