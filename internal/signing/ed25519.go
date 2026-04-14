package signing

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"

	"gokeeper/pkg/license"
)

var (
	ErrInvalidPrivatePEM = errors.New("signing: invalid private key PEM")
	ErrNotEd25519Private = errors.New("signing: private key is not Ed25519")
)

// LoadPrivateKeyFromPEM parses an Ed25519 private key from PKCS#8 PEM.
func LoadPrivateKeyFromPEM(pemBytes []byte) (ed25519.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, ErrInvalidPrivatePEM
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, ErrInvalidPrivatePEM
	}
	priv, ok := key.(ed25519.PrivateKey)
	if !ok {
		return nil, ErrNotEd25519Private
	}
	return priv, nil
}

// LoadPrivateKeyFromFile reads PEM from path.
func LoadPrivateKeyFromFile(path string) (ed25519.PrivateKey, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return LoadPrivateKeyFromPEM(b)
}

// GenerateKeyPair returns PEM-encoded PKCS#8 private key and PKIX public key.
func GenerateKeyPair() (privPEM, pubPEM []byte, err error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	privDer, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, nil, err
	}
	privPEM = pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privDer})
	pubDer, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return nil, nil, err
	}
	pubPEM = pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDer})
	return privPEM, pubPEM, nil
}

// IssueLicense creates a new license id, builds claims, and returns the wire-format key.
func IssueLicense(priv ed25519.PrivateKey, expiresAt *time.Time) (licenseID uuid.UUID, licenseKey string, err error) {
	id := uuid.New()
	c := license.Claims{LicenseID: id.String(), Exp: 0}
	if expiresAt != nil {
		c.Exp = expiresAt.Unix()
	}
	payload, err := json.Marshal(c)
	if err != nil {
		return uuid.Nil, "", err
	}
	sig := ed25519.Sign(priv, payload)
	licenseKey = base64.RawURLEncoding.EncodeToString(payload) + "." + base64.RawURLEncoding.EncodeToString(sig)
	return id, licenseKey, nil
}

// PublicPEMFromPrivate builds PKIX PEM for the public half.
func PublicPEMFromPrivate(priv ed25519.PrivateKey) ([]byte, error) {
	pub := priv.Public().(ed25519.PublicKey)
	der, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der}), nil
}

// DebugLicensePreview helps operators; avoid logging full keys in production.
func DebugLicensePreview(licenseKey string) string {
	if len(licenseKey) <= 16 {
		return "***"
	}
	return fmt.Sprintf("%s…(%d)", licenseKey[:12], len(licenseKey))
}
