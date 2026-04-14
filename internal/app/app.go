package app

import (
	"crypto/ed25519"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"gokeeper/internal/api"
	"gokeeper/internal/config"
	"gokeeper/internal/service"
	"gokeeper/internal/signing"
	"gokeeper/internal/store"
)

// LoadPrivateKey resolves the signing key from config (path, inline PEM, or dev auto-generate).
func LoadPrivateKey(cfg *config.Config) (ed25519.PrivateKey, error) {
	switch {
	case cfg.PrivateKeyPath != "":
		return signing.LoadPrivateKeyFromFile(cfg.PrivateKeyPath)
	case cfg.PrivateKeyPEM != "":
		return signing.LoadPrivateKeyFromPEM([]byte(cfg.PrivateKeyPEM))
	case cfg.GenerateKeysIfDev:
		privPath := filepath.Join(filepath.Dir(cfg.DatabasePath), "license_private.pem")
		pubPath := filepath.Join(filepath.Dir(cfg.DatabasePath), "license_public.pem")
		if _, err := os.Stat(privPath); err == nil {
			return signing.LoadPrivateKeyFromFile(privPath)
		}
		privPEM, pubPEM, err := signing.GenerateKeyPair()
		if err != nil {
			return nil, err
		}
		if err := os.WriteFile(privPath, privPEM, 0o600); err != nil {
			return nil, err
		}
		if err := os.WriteFile(pubPath, pubPEM, 0o644); err != nil {
			return nil, err
		}
		log.Printf("generated dev keys: %s and %s", privPath, pubPath)
		return signing.LoadPrivateKeyFromPEM(privPEM)
	default:
		return nil, nil
	}
}

// Build wires config, database, and HTTP routes. The returned engine is ready for Run or tests.
func Build(cfg config.Config) (*gin.Engine, error) {
	if err := config.EnsureDataDir(cfg.DatabasePath); err != nil {
		return nil, fmt.Errorf("data dir: %w", err)
	}

	priv, err := LoadPrivateKey(&cfg)
	if err != nil {
		return nil, fmt.Errorf("private key: %w", err)
	}
	if err := cfg.Validate(priv != nil); err != nil {
		return nil, err
	}

	db, err := store.Open(cfg.DatabasePath)
	if err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}
	if cfg.AutoMigrate {
		if err := store.AutoMigrate(db); err != nil {
			return nil, fmt.Errorf("migrate: %w", err)
		}
	}

	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	svc := service.NewLicense(db)
	api.Register(r, &api.Handler{Svc: svc, Priv: priv})
	return r, nil
}

// Run builds the server and listens on cfg.Addr.
func Run(cfg config.Config) error {
	r, err := Build(cfg)
	if err != nil {
		return err
	}
	log.Printf("listening on %s", cfg.Addr)
	return r.Run(cfg.Addr)
}
