// Package license provides offline validation of license keys signed with Ed25519.
package license

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrInvalidKey     = errors.New("license: invalid license key format")
	ErrInvalidPEM     = errors.New("license: invalid public key PEM")
	ErrWrongKeyType   = errors.New("license: public key is not Ed25519")
	ErrVerify         = errors.New("license: signature verification failed")
	ErrExpired        = errors.New("license: expired")
	ErrInvalidPayload = errors.New("license: invalid payload")
)

// Claims are embedded in a signed license key (JSON payload).
type Claims struct {
	LicenseID string `json:"license_id"`
	// Exp is Unix seconds; 0 means no expiry.
	Exp int64 `json:"exp"`
}

// VerifySignature checks the Ed25519 signature and unmarshals claims. It does not
// enforce expiry; use Validate for offline checks including expiry. Use this when
// the caller needs claims after signature verification (e.g. online revocation checks).
func VerifySignature(licenseKey string, publicKeyPEM []byte) (*Claims, error) {
	pub, err := parseEd25519PublicKey(publicKeyPEM)
	if err != nil {
		return nil, err
	}
	parts := strings.Split(licenseKey, ".")
	if len(parts) != 2 {
		return nil, ErrInvalidKey
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, ErrInvalidKey
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrInvalidKey
	}
	if len(sig) != ed25519.SignatureSize || !ed25519.Verify(pub, payload, sig) {
		return nil, ErrVerify
	}
	var c Claims
	if err := json.Unmarshal(payload, &c); err != nil {
		return nil, ErrInvalidPayload
	}
	if c.LicenseID == "" {
		return nil, ErrInvalidPayload
	}
	return &c, nil
}

// Validate checks the license key using the Ed25519 public key (PEM), including expiry.
func Validate(licenseKey string, publicKeyPEM []byte) (*Claims, error) {
	c, err := VerifySignature(licenseKey, publicKeyPEM)
	if err != nil {
		return nil, err
	}
	if c.Exp != 0 && time.Now().Unix() > c.Exp {
		return nil, fmt.Errorf("%w (exp %d)", ErrExpired, c.Exp)
	}
	return c, nil
}

func parseEd25519PublicKey(pemBytes []byte) (ed25519.PublicKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, ErrInvalidPEM
	}
	raw, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, ErrInvalidPEM
	}
	pub, ok := raw.(ed25519.PublicKey)
	if !ok {
		return nil, ErrWrongKeyType
	}
	return pub, nil
}
