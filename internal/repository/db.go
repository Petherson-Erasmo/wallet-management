package repository

import (
	"github.com/Petherson-Erasmo/wallet-management/internal/domain"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&domain.PortfolioItem{}, &domain.RecommendedItem{}); err != nil {
		return nil, err
	}

	return db, nil
}
