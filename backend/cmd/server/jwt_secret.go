package main

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Explicit secrets take precedence. Docker can persist an automatically generated
// secret using JWT_SECRET_FILE without sharing a default key across installations.
func loadJWTSecret() (string, error) {
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		return secret, nil
	}
	path := os.Getenv("JWT_SECRET_FILE")
	if path == "" {
		return "", fmt.Errorf("set JWT_SECRET (at least 32 bytes) or JWT_SECRET_FILE")
	}
	data, err := os.ReadFile(path)
	if err == nil {
		return strings.TrimSpace(string(data)), nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("read JWT secret file: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return "", fmt.Errorf("create JWT secret directory: %w", err)
	}
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", fmt.Errorf("generate JWT secret: %w", err)
	}
	secret := hex.EncodeToString(key)
	// Write a complete file before publishing it so concurrent starts never read a
	// partially written secret. Link refuses to overwrite another process's key.
	temporary, err := os.CreateTemp(filepath.Dir(path), ".jwt-secret-*")
	if err != nil {
		return "", err
	}
	defer os.Remove(temporary.Name())
	if _, err := temporary.WriteString(secret); err != nil {
		temporary.Close()
		return "", err
	}
	if err := temporary.Close(); err != nil {
		return "", err
	}
	if err := os.Link(temporary.Name(), path); err != nil && !errors.Is(err, os.ErrExist) {
		return "", fmt.Errorf("persist JWT secret: %w", err)
	}
	data, err = os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read persisted JWT secret: %w", err)
	}
	return strings.TrimSpace(string(data)), nil
}
