package main

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const keySize = 64

type config struct {
	KeyFile  string `json:"key_file"`
	Issuer   string `json:"issuer"`
	Audience string `json:"audience"`
}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(args []string, out io.Writer) error {
	if len(args) == 0 {
		return errors.New("usage: jwt-key-store init -key ABSOLUTE_PATH | demo -config CONFIG_PATH")
	}
	switch args[0] {
	case "init":
		flags := flag.NewFlagSet("init", flag.ContinueOnError)
		path := flags.String("key", "", "absolute path in an existing, access-restricted directory")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		if flags.NArg() != 0 || !filepath.IsAbs(*path) {
			return errors.New("init requires -key with an absolute file path and no positional arguments")
		}
		if err := generateKey(*path); err != nil {
			return err
		}
		_, err := fmt.Fprintln(out, "Created signing key (contents are not displayed).")
		return err
	case "demo":
		flags := flag.NewFlagSet("demo", flag.ContinueOnError)
		path := flags.String("config", "config.example.json", "JSON configuration path")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		if flags.NArg() != 0 {
			return errors.New("demo accepts no positional arguments")
		}
		cfg, err := loadConfig(*path)
		if err != nil {
			return err
		}
		key, err := loadKey(cfg.KeyFile)
		if err != nil {
			return err
		}
		token, err := signToken(key, cfg, "demo-user", time.Now())
		if err != nil {
			return err
		}
		claims, err := verifyToken(key, cfg, token)
		if err != nil {
			return err
		}
		// Do not print bearer tokens: possession grants access, just like a password.
		_, err = fmt.Fprintf(out, "Signed and verified an HS512 JWT for %q; expires at %s. Token not displayed.\n",
			claims.Subject, claims.ExpiresAt.Time.UTC().Format(time.RFC3339))
		return err
	default:
		return fmt.Errorf("unknown command %q (use init or demo)", args[0])
	}
}

func loadConfig(path string) (config, error) {
	var cfg config
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, fmt.Errorf("read configuration: %w", err)
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("decode configuration: %w", err)
	}
	if override := os.Getenv("JWT_KEY_FILE"); override != "" {
		cfg.KeyFile = override
	}
	if strings.TrimSpace(cfg.KeyFile) == "" || strings.TrimSpace(cfg.Issuer) == "" || strings.TrimSpace(cfg.Audience) == "" {
		return cfg, errors.New("key_file (or JWT_KEY_FILE), issuer and audience are required")
	}
	if !filepath.IsAbs(cfg.KeyFile) {
		return cfg, errors.New("key_file / JWT_KEY_FILE must be an absolute path")
	}
	return cfg, nil
}

func generateKey(path string) error {
	key := make([]byte, keySize)
	if _, err := rand.Read(key); err != nil {
		return fmt.Errorf("generate signing key: %w", err)
	}
	// O_EXCL prevents accidental rotation/overwrite of an existing key.
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return fmt.Errorf("create key file: %w", err)
	}
	_, writeErr := file.Write(key)
	if err := errors.Join(writeErr, file.Close()); err != nil {
		return fmt.Errorf("write key file: %w", errors.Join(err, os.Remove(path)))
	}
	return nil
}

func loadKey(path string) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, fmt.Errorf("inspect key file: %w", err)
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("key file must be a regular file, not a symlink or directory")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open key file: %w", err)
	}
	defer file.Close()
	opened, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("inspect opened key file: %w", err)
	}
	if !os.SameFile(info, opened) {
		return nil, errors.New("key file changed while opening")
	}
	// Windows mode bits do not describe DACLs; see the mandatory ACL setup in README.
	if runtime.GOOS != "windows" && opened.Mode().Perm()&0077 != 0 {
		return nil, errors.New("key file must not grant group or other permissions (use chmod 600)")
	}
	key, err := io.ReadAll(io.LimitReader(file, keySize+1))
	if err != nil {
		return nil, fmt.Errorf("read key file: %w", err)
	}
	if len(key) != keySize {
		return nil, fmt.Errorf("key file must contain exactly %d raw random bytes; run init", keySize)
	}
	return key, nil
}

func signToken(key []byte, cfg config, subject string, now time.Time) (string, error) {
	if len(key) != keySize {
		return "", errors.New("HS512 requires a 64-byte key in this example")
	}
	claims := jwt.RegisteredClaims{
		Issuer:    cfg.Issuer,
		Audience:  jwt.ClaimStrings{cfg.Audience},
		Subject:   subject,
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(5 * time.Minute)),
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS512, claims).SignedString(key)
	if err != nil {
		return "", fmt.Errorf("sign JWT: %w", err)
	}
	return token, nil
}

func verifyToken(key []byte, cfg config, raw string) (*jwt.RegisteredClaims, error) {
	if len(key) != keySize {
		return nil, errors.New("HS512 requires a 64-byte key in this example")
	}
	claims := &jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(raw, claims, func(_ *jwt.Token) (interface{}, error) {
		return key, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS512.Alg()}),
		jwt.WithIssuer(cfg.Issuer), jwt.WithAudience(cfg.Audience),
		jwt.WithExpirationRequired(), jwt.WithIssuedAt())
	if err != nil {
		return nil, fmt.Errorf("verify JWT: %w", err)
	}
	if claims.Subject == "" {
		return nil, errors.New("verify JWT: subject is required")
	}
	return claims, nil
}
