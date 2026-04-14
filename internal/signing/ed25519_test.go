package signing

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gokeeper/pkg/license"
)

func TestGenerateKeyPair_roundTrip(t *testing.T) {
	privPEM, pubPEM, err := GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	priv, err := LoadPrivateKeyFromPEM(privPEM)
	if err != nil {
		t.Fatal(err)
	}
	if len(priv) != ed25519.PrivateKeySize {
		t.Fatalf("priv len %d", len(priv))
	}
	outPub, err := PublicPEMFromPrivate(priv)
	if err != nil {
		t.Fatal(err)
	}
	if string(outPub) != string(pubPEM) {
		t.Fatal("public PEM mismatch")
	}
}

func TestLoadPrivateKeyFromPEM_invalid(t *testing.T) {
	_, err := LoadPrivateKeyFromPEM([]byte("not pem"))
	if err != ErrInvalidPrivatePEM {
		t.Fatalf("got %v", err)
	}
	badDER := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: []byte{0x01}})
	_, err = LoadPrivateKeyFromPEM(badDER)
	if err != ErrInvalidPrivatePEM {
		t.Fatalf("got %v", err)
	}
}

func TestLoadPrivateKeyFromFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "key.pem")
	_, err := LoadPrivateKeyFromFile(path)
	if err == nil {
		t.Fatal("expected error for missing file")
	}

	privPEM, _, err := GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, privPEM, 0o600); err != nil {
		t.Fatal(err)
	}
	priv, err := LoadPrivateKeyFromFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(priv) != ed25519.PrivateKeySize {
		t.Fatal("bad key")
	}
}

func TestIssueLicense(t *testing.T) {
	_, priv, _ := ed25519.GenerateKey(rand.Reader)
	id, key, err := IssueLicense(priv, nil)
	if err != nil || id.String() == "" || key == "" {
		t.Fatalf("issue: %v %s %s", err, id, key)
	}
	pubPEM, err := PublicPEMFromPrivate(priv)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := license.Validate(key, pubPEM); err != nil {
		t.Fatal(err)
	}

	exp := time.Now().UTC().Add(time.Hour)
	id2, key2, err := IssueLicense(priv, &exp)
	if err != nil {
		t.Fatal(err)
	}
	c, err := license.VerifySignature(key2, pubPEM)
	if err != nil || c.Exp != exp.Unix() {
		t.Fatalf("exp claim: %+v err %v", c, err)
	}
	_ = id2
}

func TestDebugLicensePreview(t *testing.T) {
	if p := DebugLicensePreview("short"); p != "***" {
		t.Fatalf("%q", p)
	}
	if p := DebugLicensePreview("longer_than_sixteen_chars"); p == "***" || len(p) < 10 {
		t.Fatalf("%q", p)
	}
}
