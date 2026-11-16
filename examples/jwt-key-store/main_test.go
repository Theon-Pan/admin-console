package main

import (
	"bytes"
	"crypto/rand"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestKeyLifecycle(t *testing.T) {
	path := filepath.Join(t.TempDir(), "signing.key")
	if _, err := loadKey(path); err == nil {
		t.Fatal("missing key accepted")
	}
	if err := generateKey(path); err != nil {
		t.Fatal(err)
	}
	key, err := loadKey(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(key) != keySize {
		t.Fatalf("key length = %d", len(key))
	}
	if err := generateKey(path); err == nil {
		t.Fatal("existing key overwritten")
	}
	again, err := loadKey(path)
	if err != nil || !bytes.Equal(key, again) {
		t.Fatal("key did not persist unchanged", err)
	}
}

func TestRejectMalformedKeys(t *testing.T) {
	for _, size := range []int{0, 32, 63, 65, 4096} {
		path := filepath.Join(t.TempDir(), "bad.key")
		if err := os.WriteFile(path, make([]byte, size), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := loadKey(path); err == nil {
			t.Errorf("accepted %d-byte key", size)
		}
	}
	if _, err := loadKey(t.TempDir()); err == nil {
		t.Fatal("accepted directory as key")
	}
}

func TestRejectInsecurePermissionsAndSymlinks(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows requires DACL setup; POSIX permissions do not apply")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "signing.key")
	if err := generateKey(path); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "link.key")
	if err := os.Symlink(path, link); err != nil {
		t.Fatal(err)
	}
	if _, err := loadKey(link); err == nil {
		t.Fatal("accepted symlink")
	}
	if err := os.Chmod(path, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := loadKey(path); err == nil {
		t.Fatal("accepted world-readable key")
	}
}

func TestTokenValidation(t *testing.T) {
	key := make([]byte, keySize)
	if _, err := rand.Read(key); err != nil {
		t.Fatal(err)
	}
	cfg := config{Issuer: "test-issuer", Audience: "test-audience"}
	now := time.Now().Truncate(time.Second)
	raw, err := signToken(key, cfg, "user-123", now)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := verifyToken(key, cfg, raw)
	if err != nil || claims.Subject != "user-123" {
		t.Fatal("valid token rejected", err)
	}
	if claims.ExpiresAt.Sub(claims.IssuedAt.Time) != 5*time.Minute {
		t.Fatal("unexpected token lifetime")
	}
	tests := []struct {
		name   string
		method jwt.SigningMethod
		change func(*jwt.RegisteredClaims)
	}{
		{"wrong algorithm", jwt.SigningMethodHS256, func(*jwt.RegisteredClaims) {}},
		{"wrong issuer", jwt.SigningMethodHS512, func(c *jwt.RegisteredClaims) { c.Issuer = "other" }},
		{"wrong audience", jwt.SigningMethodHS512, func(c *jwt.RegisteredClaims) { c.Audience = jwt.ClaimStrings{"other"} }},
		{"missing expiry", jwt.SigningMethodHS512, func(c *jwt.RegisteredClaims) { c.ExpiresAt = nil }},
		{"expired", jwt.SigningMethodHS512, func(c *jwt.RegisteredClaims) { c.ExpiresAt = jwt.NewNumericDate(now.Add(-time.Minute)) }},
		{"future not-before", jwt.SigningMethodHS512, func(c *jwt.RegisteredClaims) { c.NotBefore = jwt.NewNumericDate(now.Add(time.Hour)) }},
		{"future issued-at", jwt.SigningMethodHS512, func(c *jwt.RegisteredClaims) { c.IssuedAt = jwt.NewNumericDate(now.Add(time.Hour)) }},
		{"missing subject", jwt.SigningMethodHS512, func(c *jwt.RegisteredClaims) { c.Subject = "" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changed := *claims
			tt.change(&changed)
			bad, err := jwt.NewWithClaims(tt.method, changed).SignedString(key)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := verifyToken(key, cfg, bad); err == nil {
				t.Fatal("invalid token accepted")
			}
		})
	}
	otherKey := append([]byte(nil), key...)
	otherKey[0] ^= 0xff
	if _, err := verifyToken(otherKey, cfg, raw); err == nil {
		t.Fatal("accepted wrong signing key")
	}
	parts := strings.Split(raw, ".")
	parts[1] = "eyJzdWIiOiJhdHRhY2tlciJ9"
	if _, err := verifyToken(key, cfg, strings.Join(parts, ".")); err == nil {
		t.Fatal("accepted tampered payload")
	}
}

func TestConfigAndDemo(t *testing.T) {
	dir := t.TempDir()
	keyPath := filepath.Join(dir, "signing.key")
	t.Setenv("JWT_KEY_FILE", "")
	configPath := filepath.Join(dir, "config.json")
	if err := os.WriteFile(configPath, []byte(`{"issuer":"test","audience":"demo"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadConfig(configPath); err == nil {
		t.Fatal("accepted missing key path")
	}
	t.Setenv("JWT_KEY_FILE", "relative.key")
	if _, err := loadConfig(configPath); err == nil {
		t.Fatal("accepted relative key path")
	}
	t.Setenv("JWT_KEY_FILE", keyPath)
	var out bytes.Buffer
	if err := run([]string{"init", "-key", keyPath}, &out); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"demo", "-config", configPath}, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), `JWT for "demo-user"`) {
		t.Fatalf("unexpected demo output: %s", out.String())
	}
	key, err := loadKey(keyPath)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(out.Bytes(), key) || strings.Contains(out.String(), "eyJ") {
		t.Fatal("output leaked key or token")
	}
}
