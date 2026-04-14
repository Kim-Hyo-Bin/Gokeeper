package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestFromEnv_defaults(t *testing.T) {
	t.Setenv("ADDR", "")
	t.Setenv("DATABASE_PATH", "")
	t.Setenv("LICENSE_PRIVATE_KEY_PATH", "")
	t.Setenv("LICENSE_PRIVATE_KEY_PEM", "")
	t.Setenv("AUTO_MIGRATE", "")
	t.Setenv("LICENSE_GENERATE_KEYS_DEV", "")

	c := FromEnv()
	if c.Addr != ":8080" || c.DatabasePath != "./data/licenses.db" {
		t.Fatalf("%+v", c)
	}
	if !c.AutoMigrate || c.GenerateKeysIfDev {
		t.Fatalf("%+v", c)
	}
}

func TestFromEnv_overrides(t *testing.T) {
	t.Setenv("ADDR", ":9090")
	t.Setenv("DATABASE_PATH", "/tmp/x.db")
	t.Setenv("LICENSE_PRIVATE_KEY_PATH", "/keys/priv.pem")
	t.Setenv("LICENSE_PRIVATE_KEY_PEM", "PEM")
	t.Setenv("AUTO_MIGRATE", "false")
	t.Setenv("LICENSE_GENERATE_KEYS_DEV", "true")

	c := FromEnv()
	if c.Addr != ":9090" || c.DatabasePath != "/tmp/x.db" {
		t.Fatal()
	}
	if c.PrivateKeyPath != "/keys/priv.pem" || c.PrivateKeyPEM != "PEM" {
		t.Fatal()
	}
	if c.AutoMigrate || !c.GenerateKeysIfDev {
		t.Fatal()
	}
}

func TestFromEnv_boolInvalidFallsBack(t *testing.T) {
	t.Setenv("AUTO_MIGRATE", "not-a-bool")
	c := FromEnv()
	if !c.AutoMigrate {
		t.Fatal("expected default true on parse error")
	}
}

func TestEnsureDataDir(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "nested", "db.sqlite")
	if err := EnsureDataDir(dbPath); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "nested")); errors.Is(err, os.ErrNotExist) {
		t.Fatal("dir not created")
	}
	if err := EnsureDataDir("bare.db"); err != nil {
		t.Fatal(err)
	}
}

func TestConfig_Validate(t *testing.T) {
	var c Config
	if err := c.Validate(true); err != nil {
		t.Fatal(err)
	}
	if err := c.Validate(false); err == nil {
		t.Fatal("expected error")
	}
}
