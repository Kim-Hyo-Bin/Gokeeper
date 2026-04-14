package model

import (
	"time"

	"github.com/google/uuid"
)

// License is persisted metadata for issued keys (UUID + optional expiry/revocation).
type License struct {
	ID         uuid.UUID  `gorm:"type:text;primaryKey" json:"id"`
	LicenseKey string     `gorm:"type:text;not null" json:"-"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

func (License) TableName() string {
	return "licenses"
}
