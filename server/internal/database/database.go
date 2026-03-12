package database

import (
	"agent-tracker/internal/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init(dataSourceName string) error {
	var err error
	DB, err = gorm.Open(sqlite.Open(dataSourceName), &gorm.Config{})
	if err != nil {
		return err
	}

	return DB.AutoMigrate(&models.Tool{}, &models.Entry{}, &models.SyncState{})
}
