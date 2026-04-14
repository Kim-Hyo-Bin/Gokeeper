package license

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"testing"
	"time"
)

func TestValidate_roundTrip(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	pemBytes := mustPublicPEM(t, pub)
	claims := Claims{LicenseID: "550e8400-e29b-41d4-a716-446655440000", Exp: time.Now().Add(time.Hour).Unix()}
	payload, _ := json.Marshal(claims)
	sig := ed25519.Sign(priv, payload)
	key := encodeLicenseKey(payload, sig)

	got, err := Validate(key, pemBytes)
	if err != nil {
		t.Fatal(err)
	}
	if got.LicenseID != claims.LicenseID || got.Exp != claims.Exp {
		t.Fatalf("claims mismatch: %+v vs %+v", got, claims)
	}
}

func TestValidate_expired(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	pemBytes := mustPublicPEM(t, pub)
	claims := Claims{LicenseID: "550e8400-e29b-41d4-a716-446655440000", Exp: time.Now().Add(-time.Hour).Unix()}
	payload, _ := json.Marshal(claims)
	sig := ed25519.Sign(priv, payload)
	key := encodeLicenseKey(payload, sig)
	_, err := Validate(key, pemBytes)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestVerifySignature_wrongPublicKeyType(t *testing.T) {
	_, priv, _ := ed25519.GenerateKey(rand.Reader)
	claims := Claims{LicenseID: "550e8400-e29b-41d4-a716-446655440000", Exp: 0}
	payload, _ := json.Marshal(claims)
	sig := ed25519.Sign(priv, payload)
	key := encodeLicenseKey(payload, sig)

	ecPriv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	ecDer, _ := x509.MarshalPKIXPublicKey(&ecPriv.PublicKey)
	ecPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: ecDer})
	_, err := VerifySignature(key, ecPEM)
	if !errors.Is(err, ErrWrongKeyType) {
		t.Fatalf("got %v", err)
	}
}

func TestVerifySignature_invalidPEM(t *testing.T) {
	_, priv, _ := ed25519.GenerateKey(rand.Reader)
	claims := Claims{LicenseID: "550e8400-e29b-41d4-a716-446655440000", Exp: 0}
	payload, _ := json.Marshal(claims)
	sig := ed25519.Sign(priv, payload)
	key := encodeLicenseKey(payload, sig)
	_, err := VerifySignature(key, []byte("not pem"))
	if !errors.Is(err, ErrInvalidPEM) {
		t.Fatalf("got %v", err)
	}
}

func TestVerifySignature_ignoresExpiry(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	pemBytes := mustPublicPEM(t, pub)
	claims := Claims{LicenseID: "550e8400-e29b-41d4-a716-446655440000", Exp: time.Now().Add(-time.Hour).Unix()}
	payload, _ := json.Marshal(claims)
	sig := ed25519.Sign(priv, payload)
	key := encodeLicenseKey(payload, sig)
	got, err := VerifySignature(key, pemBytes)
	if err != nil {
		t.Fatal(err)
	}
	if got.LicenseID != claims.LicenseID || got.Exp != claims.Exp {
		t.Fatalf("claims: %+v vs %+v", got, claims)
	}
}

func encodeLicenseKey(payload, sig []byte) string {
	return base64.RawURLEncoding.EncodeToString(payload) + "." + base64.RawURLEncoding.EncodeToString(sig)
}

func mustPublicPEM(t *testing.T, pub ed25519.PublicKey) []byte {
	t.Helper()
	der, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		t.Fatal(err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})
}
