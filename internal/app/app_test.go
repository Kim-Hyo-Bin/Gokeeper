package app

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"gokeeper/internal/config"
	"gokeeper/internal/signing"
)

func TestLoadPrivateKey_fromPEM(t *testing.T) {
	privPEM, _, err := signing.GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{PrivateKeyPEM: string(privPEM)}
	priv, err := LoadPrivateKey(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(priv) != 64 {
		t.Fatalf("key len %d", len(priv))
	}
}

func TestLoadPrivateKey_fromFile(t *testing.T) {
	dir := t.TempDir()
	privPEM, _, err := signing.GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "k.pem")
	if err := os.WriteFile(path, privPEM, 0o600); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{PrivateKeyPath: path}
	priv, err := LoadPrivateKey(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(priv) != 64 {
		t.Fatal("bad key")
	}
}

func TestLoadPrivateKey_devGeneratesPair(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "nested", "db.sqlite")
	cfg := &config.Config{
		DatabasePath:      dbPath,
		GenerateKeysIfDev: true,
	}
	if err := config.EnsureDataDir(dbPath); err != nil {
		t.Fatal(err)
	}
	priv, err := LoadPrivateKey(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(priv) != 64 {
		t.Fatal("bad key")
	}
	priv2, err := LoadPrivateKey(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if string(priv) != string(priv2) {
		t.Fatal("expected same key from existing file")
	}
}

func TestLoadPrivateKey_none(t *testing.T) {
	cfg := &config.Config{}
	priv, err := LoadPrivateKey(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if priv != nil {
		t.Fatal("expected nil")
	}
}

func TestBuild_health(t *testing.T) {
	t.Setenv("GIN_MODE", "test")
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "db.sqlite")
	privPEM, _, err := signing.GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{
		Addr:          ":0",
		DatabasePath:  dbPath,
		PrivateKeyPEM: string(privPEM),
		AutoMigrate:   true,
	}
	r, err := Build(cfg)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/health", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}
}

func TestBuild_validationError(t *testing.T) {
	t.Setenv("GIN_MODE", "test")
	cfg := config.Config{
		DatabasePath: filepath.Join(t.TempDir(), "x.db"),
		// no key, no dev flag
	}
	_, err := Build(cfg)
	if err == nil {
		t.Fatal("expected error")
	}
}
