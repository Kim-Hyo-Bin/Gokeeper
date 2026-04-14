package service

import (
	"context"
	"crypto/ed25519"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"gokeeper/internal/model"
	"gokeeper/internal/signing"
	"gokeeper/pkg/license"
)

// ErrLicenseNotFound is returned when no row matches the license id for revoke.
var ErrLicenseNotFound = errors.New("service: license not found")

// License coordinates signing, persistence, and online verification.
type License struct {
	db *gorm.DB
}

// NewLicense builds a license service backed by db. The caller must supply a
// configured *gorm.DB (migrations applied as needed).
func NewLicense(db *gorm.DB) *License {
	return &License{db: db}
}

// Issue creates a signed license and stores it.
func (s *License) Issue(ctx context.Context, priv ed25519.PrivateKey, expiresAt *time.Time) (id uuid.UUID, licenseKey string, err error) {
	id, licenseKey, err = signing.IssueLicense(priv, expiresAt)
	if err != nil {
		return uuid.Nil, "", err
	}
	rec := model.License{
		ID:         id,
		LicenseKey: licenseKey,
		ExpiresAt:  expiresAt,
		CreatedAt:  time.Now().UTC(),
	}
	if err := s.db.WithContext(ctx).Create(&rec).Error; err != nil {
		return uuid.Nil, "", err
	}
	return id, licenseKey, nil
}

// Get returns a license row by id.
func (s *License) Get(ctx context.Context, id uuid.UUID) (model.License, error) {
	var rec model.License
	err := s.db.WithContext(ctx).First(&rec, "id = ?", id).Error
	return rec, err
}

// Revoke sets revoked_at for the given license id.
func (s *License) Revoke(ctx context.Context, id uuid.UUID) (revokedAt time.Time, err error) {
	now := time.Now().UTC()
	res := s.db.WithContext(ctx).Model(&model.License{}).Where("id = ?", id).Update("revoked_at", now)
	if res.Error != nil {
		return time.Time{}, res.Error
	}
	if res.RowsAffected == 0 {
		return time.Time{}, ErrLicenseNotFound
	}
	return now, nil
}

// VerifyOutcome is the result of an online verification against this server's key and DB.
type VerifyOutcome struct {
	Valid     bool       `json:"valid"`
	Reason    string     `json:"reason"`
	LicenseID *uuid.UUID `json:"license_id,omitempty"`
	Exp       int64      `json:"exp,omitempty"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
}

// Verify checks the Ed25519 signature with the server public key derived from priv,
// then checks the database for revocation and row integrity.
func (s *License) Verify(ctx context.Context, priv ed25519.PrivateKey, licenseKey string) (VerifyOutcome, error) {
	pubPEM, err := signing.PublicPEMFromPrivate(priv)
	if err != nil {
		return VerifyOutcome{}, err
	}
	claims, err := license.VerifySignature(licenseKey, pubPEM)
	if err != nil {
		return mapCryptoError(err), nil
	}
	id, err := uuid.Parse(claims.LicenseID)
	if err != nil {
		return VerifyOutcome{Valid: false, Reason: "invalid_claims"}, nil
	}
	expired := claims.Exp != 0 && time.Now().Unix() > claims.Exp

	rec, err := s.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return VerifyOutcome{
				Valid:     false,
				Reason:    "unknown",
				LicenseID: &id,
				Exp:       claims.Exp,
			}, nil
		}
		return VerifyOutcome{}, err
	}
	if rec.LicenseKey != licenseKey {
		return VerifyOutcome{
			Valid:     false,
			Reason:    "mismatch",
			LicenseID: &id,
			Exp:       claims.Exp,
		}, nil
	}
	if rec.RevokedAt != nil {
		ra := *rec.RevokedAt
		return VerifyOutcome{
			Valid:     false,
			Reason:    "revoked",
			LicenseID: &id,
			Exp:       claims.Exp,
			RevokedAt: &ra,
			ExpiresAt: rec.ExpiresAt,
			CreatedAt: &rec.CreatedAt,
		}, nil
	}
	if expired {
		return VerifyOutcome{
			Valid:     false,
			Reason:    "expired",
			LicenseID: &id,
			Exp:       claims.Exp,
			ExpiresAt: rec.ExpiresAt,
			CreatedAt: &rec.CreatedAt,
		}, nil
	}
	return VerifyOutcome{
		Valid:     true,
		Reason:    "ok",
		LicenseID: &id,
		Exp:       claims.Exp,
		ExpiresAt: rec.ExpiresAt,
		CreatedAt: &rec.CreatedAt,
	}, nil
}

func mapCryptoError(err error) VerifyOutcome {
	switch {
	case errors.Is(err, license.ErrVerify):
		return VerifyOutcome{Valid: false, Reason: "invalid_signature"}
	case errors.Is(err, license.ErrInvalidKey):
		return VerifyOutcome{Valid: false, Reason: "invalid_format"}
	case errors.Is(err, license.ErrInvalidPayload):
		return VerifyOutcome{Valid: false, Reason: "invalid_payload"}
	case errors.Is(err, license.ErrInvalidPEM), errors.Is(err, license.ErrWrongKeyType):
		return VerifyOutcome{Valid: false, Reason: "server_key_error"}
	default:
		return VerifyOutcome{Valid: false, Reason: "invalid_format"}
	}
}
