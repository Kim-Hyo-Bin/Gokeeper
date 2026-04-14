package service

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"gokeeper/internal/model"
	"gokeeper/internal/signing"
	"gokeeper/pkg/license"
)

func testDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.License{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestLicense_IssueGetRevokeVerify(t *testing.T) {
	ctx := context.Background()
	db := testDB(t)
	svc := NewLicense(db)
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	exp := time.Now().UTC().Add(time.Hour)
	id, key, err := svc.Issue(ctx, priv, &exp)
	if err != nil {
		t.Fatal(err)
	}
	if id == uuid.Nil || key == "" {
		t.Fatal("expected id and key")
	}

	rec, err := svc.Get(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if rec.LicenseKey != key {
		t.Fatal("stored key mismatch")
	}

	pubPEM, err := signing.PublicPEMFromPrivate(priv)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := license.Validate(key, pubPEM); err != nil {
		t.Fatal(err)
	}

	out, err := svc.Verify(ctx, priv, key)
	if err != nil {
		t.Fatal(err)
	}
	if !out.Valid || out.Reason != "ok" {
		t.Fatalf("verify: %+v", out)
	}

	if _, err := svc.Revoke(ctx, id); err != nil {
		t.Fatal(err)
	}

	out2, err := svc.Verify(ctx, priv, key)
	if err != nil {
		t.Fatal(err)
	}
	if out2.Valid || out2.Reason != "revoked" {
		t.Fatalf("after revoke: %+v", out2)
	}
}

func TestLicense_Verify_expired(t *testing.T) {
	ctx := context.Background()
	db := testDB(t)
	svc := NewLicense(db)
	_, priv, _ := ed25519.GenerateKey(rand.Reader)
	past := time.Now().UTC().Add(-time.Hour)
	_, key, err := svc.Issue(ctx, priv, &past)
	if err != nil {
		t.Fatal(err)
	}
	out, err := svc.Verify(ctx, priv, key)
	if err != nil {
		t.Fatal(err)
	}
	if out.Valid || out.Reason != "expired" {
		t.Fatalf("got %+v", out)
	}
}

func TestLicense_Revoke_notFound(t *testing.T) {
	ctx := context.Background()
	db := testDB(t)
	svc := NewLicense(db)
	_, err := svc.Revoke(ctx, uuid.New())
	if !errors.Is(err, ErrLicenseNotFound) {
		t.Fatalf("expected ErrLicenseNotFound, got %v", err)
	}
}

func TestLicense_Verify_invalidSignature(t *testing.T) {
	ctx := context.Background()
	db := testDB(t)
	svc := NewLicense(db)
	_, priv, _ := ed25519.GenerateKey(rand.Reader)
	// Two valid base64url segments, signature length not Ed25519 size -> invalid_signature
	out, err := svc.Verify(ctx, priv, "e30.e30")
	if err != nil {
		t.Fatal(err)
	}
	if out.Valid || out.Reason != "invalid_signature" {
		t.Fatalf("%+v", out)
	}
}

func TestLicense_Verify_mismatch(t *testing.T) {
	ctx := context.Background()
	db := testDB(t)
	svc := NewLicense(db)
	_, priv, _ := ed25519.GenerateKey(rand.Reader)
	id, key, err := svc.Issue(ctx, priv, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&model.License{}).Where("id = ?", id).Update("license_key", key+"tampered").Error; err != nil {
		t.Fatal(err)
	}
	out, err := svc.Verify(ctx, priv, key)
	if err != nil {
		t.Fatal(err)
	}
	if out.Valid || out.Reason != "mismatch" {
		t.Fatalf("%+v", out)
	}
}

func TestLicense_Verify_unknown(t *testing.T) {
	ctx := context.Background()
	db := testDB(t)
	svc := NewLicense(db)
	_, priv, _ := ed25519.GenerateKey(rand.Reader)
	id := uuid.New()
	payload, err := json.Marshal(license.Claims{LicenseID: id.String(), Exp: 0})
	if err != nil {
		t.Fatal(err)
	}
	sig := ed25519.Sign(priv, payload)
	key := base64.RawURLEncoding.EncodeToString(payload) + "." + base64.RawURLEncoding.EncodeToString(sig)

	out, err := svc.Verify(ctx, priv, key)
	if err != nil {
		t.Fatal(err)
	}
	if out.Valid || out.Reason != "unknown" {
		t.Fatalf("got %+v", out)
	}
}
