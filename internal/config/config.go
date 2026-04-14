package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds server runtime settings (container-friendly via env).
type Config struct {
	Addr              string
	DatabasePath      string
	PrivateKeyPath    string
	PrivateKeyPEM     string // optional inline PEM (e.g. from secret env); path wins if both set
	AutoMigrate       bool
	GenerateKeysIfDev bool // if true and no key, write keys to data dir (dev only)
}

func FromEnv() Config {
	return Config{
		Addr:              getEnv("ADDR", ":8080"),
		DatabasePath:      getEnv("DATABASE_PATH", "./data/licenses.db"),
		PrivateKeyPath:    os.Getenv("LICENSE_PRIVATE_KEY_PATH"),
		PrivateKeyPEM:     os.Getenv("LICENSE_PRIVATE_KEY_PEM"),
		AutoMigrate:       getEnvBool("AUTO_MIGRATE", true),
		GenerateKeysIfDev: getEnvBool("LICENSE_GENERATE_KEYS_DEV", false),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvBool(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

// EnsureDataDir creates parent dir for SQLite file if needed.
func EnsureDataDir(path string) error {
	dir := path
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			dir = path[:i]
			break
		}
	}
	if dir == "" || dir == path {
		return nil
	}
	return os.MkdirAll(dir, 0o750)
}

// Validate checks required fields after key resolution.
func (c *Config) Validate(hasPrivateKey bool) error {
	if !hasPrivateKey {
		return fmt.Errorf("LICENSE_PRIVATE_KEY_PATH or LICENSE_PRIVATE_KEY_PEM is required (or set LICENSE_GENERATE_KEYS_DEV=true for local dev)")
	}
	return nil
}
