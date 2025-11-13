package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FileAsset struct {
	ID               string    `gorm:"primaryKey;size:36"`
	OriginalName     string    `gorm:"size:255;not null"`
	StoredPath       string    `gorm:"size:512;uniqueIndex;not null"`
	UserID           *string   `gorm:"size:64;index"`
	ContentType      string    `gorm:"size:128;not null"`
	SizeBytes        int64     `gorm:"not null"`
	EncryptionAlg    string    `gorm:"size:32;not null;default:'AES-256'"`
	AuthenticationAlg string   `gorm:"size:32;not null;default:'GCM'"`
	CreatedAt        time.Time `gorm:"autoCreateTime;not null"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime;not null"`
	DeletedAt        gorm.DeletedAt `gorm:"index"`
}

func (f *FileAsset) BeforeCreate(tx *gorm.DB) error {
	if f.ID == "" {
		f.ID = uuid.NewString()
	}
	return nil
}

func (FileAsset) TableName() string {
	return "file_assets"
}

