package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ExcelExport struct {
	ID         string         `gorm:"primaryKey;size:36"`
	StoredPath string         `gorm:"size:512;uniqueIndex;not null"`
	CreatedAt  time.Time      `gorm:"autoCreateTime;not null"`
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}

func (e *ExcelExport) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuid.NewString()
	}
	return nil
}

func (ExcelExport) TableName() string {
	return "excel_exports"
}

