package api

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"gokeeper/internal/model"
	"gokeeper/internal/service"
)

func testRouter(t *testing.T) (*gin.Engine, ed25519.PrivateKey) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.License{}); err != nil {
		t.Fatal(err)
	}
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	svc := service.NewLicense(db)
	h := &Handler{Svc: svc, Priv: priv}
	r := gin.New()
	Register(r, h)
	return r, priv
}

func TestAPI_issueGetVerifyRevoke(t *testing.T) {
	r, _ := testRouter(t)

	// issue
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/licenses", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("issue status %d %s", w.Code, w.Body.String())
	}
	var issue struct {
		ID         uuid.UUID `json:"id"`
		LicenseKey string    `json:"license_key"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &issue); err != nil {
		t.Fatal(err)
	}

	// get
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/v1/licenses/"+issue.ID.String(), nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("get %d", w.Code)
	}

	// verify ok
	body, _ := json.Marshal(map[string]string{"license_key": issue.LicenseKey})
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/v1/licenses/verify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("verify %d %s", w.Code, w.Body.String())
	}
	var v struct {
		Valid  bool   `json:"valid"`
		Reason string `json:"reason"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &v); err != nil {
		t.Fatal(err)
	}
	if !v.Valid || v.Reason != "ok" {
		t.Fatalf("%+v", v)
	}

	// revoke
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/v1/licenses/"+issue.ID.String()+"/revoke", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("revoke %d", w.Code)
	}

	// verify revoked
	body, _ = json.Marshal(map[string]string{"license_key": issue.LicenseKey})
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/v1/licenses/verify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if err := json.Unmarshal(w.Body.Bytes(), &v); err != nil {
		t.Fatal(err)
	}
	if v.Valid || v.Reason != "revoked" {
		t.Fatalf("%+v", v)
	}
}

func TestAPI_verifyInvalidBody(t *testing.T) {
	r, _ := testRouter(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/licenses/verify", strings.NewReader(`{`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("%d", w.Code)
	}
}

func TestAPI_revokeNotFound(t *testing.T) {
	r, _ := testRouter(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/licenses/"+uuid.New().String()+"/revoke", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("%d", w.Code)
	}
}

func TestAPI_getInvalidID(t *testing.T) {
	r, _ := testRouter(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/licenses/not-uuid", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("%d", w.Code)
	}
}

func TestAPI_issueWithExpiry(t *testing.T) {
	r, _ := testRouter(t)
	payload, err := json.Marshal(map[string]string{
		"expires_at": time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339Nano),
	})
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/licenses", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("%d %s", w.Code, w.Body.String())
	}
}

func TestHealth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/health", Health)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/health", nil))
	if w.Code != http.StatusOK {
		t.Fatal(w.Code)
	}
}
