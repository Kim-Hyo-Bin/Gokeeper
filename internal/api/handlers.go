package api

import (
	"crypto/ed25519"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"gokeeper/internal/service"
)

// Health-check easter egg: innocuous query key must match healthEggQueryValue; optional `p`
// carries an arbitrary string echoed under a bland JSON field (omit `p` to echo the trigger value).
const (
	healthQueryKey   = "author"
	healthQueryValue = "kim-hyo-bin"
	healthPayloadKey = "p"
	healthValueField = "Build Data"
)

// Handler wires HTTP to the license service.
type Handler struct {
	Svc  *service.License
	Priv ed25519.PrivateKey
}

type issueRequest struct {
	ExpiresAt *time.Time `json:"expires_at"`
}

type issueResponse struct {
	ID         uuid.UUID  `json:"id"`
	LicenseKey string     `json:"license_key"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
}

type verifyRequest struct {
	LicenseKey string `json:"license_key" binding:"required"`
}

// IssueLicense POST /v1/licenses
func (h *Handler) IssueLicense(c *gin.Context) {
	var req issueRequest
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}
	id, key, err := h.Svc.Issue(c.Request.Context(), h.Priv, req.ExpiresAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, issueResponse{ID: id, LicenseKey: key, ExpiresAt: req.ExpiresAt})
}

// RevokeLicense POST /v1/licenses/:id/revoke
func (h *Handler) RevokeLicense(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	revokedAt, err := h.Svc.Revoke(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrLicenseNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id, "revoked_at": revokedAt})
}

// GetLicense GET /v1/licenses/:id
func (h *Handler) GetLicense(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	rec, err := h.Svc.Get(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":          rec.ID,
		"expires_at":  rec.ExpiresAt,
		"revoked_at":  rec.RevokedAt,
		"created_at":  rec.CreatedAt,
		"license_key": rec.LicenseKey,
	})
}

// VerifyLicense POST /v1/licenses/verify
func (h *Handler) VerifyLicense(c *gin.Context) {
	var req verifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	out, err := h.Svc.Verify(c.Request.Context(), h.Priv, req.LicenseKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, out)
}

// Health GET /health
func Health(c *gin.Context) {
	body := gin.H{"status": "ok"}
	c.JSON(http.StatusOK, body)
}
