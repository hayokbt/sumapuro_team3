package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadJWTSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "")
	t.Setenv("JWT_SECRET_FILE", "")
	if _, err := loadJWTSecret(); err == nil {
		t.Fatal("missing configuration accepted")
	}
	path := filepath.Join(t.TempDir(), "auth", "jwt-secret")
	t.Setenv("JWT_SECRET_FILE", path)
	first, err := loadJWTSecret()
	if err != nil || len(first) != 64 {
		t.Fatalf("generation: length=%d error=%v", len(first), err)
	}
	second, err := loadJWTSecret()
	if err != nil || first != second {
		t.Fatal("secret changed on restart", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatal("secret must have mode 0600")
	}
	t.Setenv("JWT_SECRET", "explicit-key-takes-precedence-123456")
	explicit, err := loadJWTSecret()
	if err != nil || explicit != os.Getenv("JWT_SECRET") {
		t.Fatal("explicit secret ignored", err)
	}
}
