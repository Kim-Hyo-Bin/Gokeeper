package store

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"gokeeper/internal/model"
)

func TestOpen_AutoMigrate(t *testing.T) {
	db, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	if err := AutoMigrate(db); err != nil {
		t.Fatal(err)
	}
	rec := model.License{ID: uuid.New(), LicenseKey: "k", CreatedAt: time.Now().UTC()}
	if err := db.Create(&rec).Error; err != nil {
		t.Fatal(err)
	}
	var got model.License
	if err := db.First(&got, "id = ?", rec.ID).Error; err != nil {
		t.Fatal(err)
	}
}
