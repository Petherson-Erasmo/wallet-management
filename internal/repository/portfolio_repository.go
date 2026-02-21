package repository

import (
	"github.com/Petherson-Erasmo/wallet-management/internal/domain"
	"gorm.io/gorm"
)

type PortfolioRepository struct {
	db *gorm.DB
}

func NewPortfolioRepository(db *gorm.DB) *PortfolioRepository {
	return &PortfolioRepository{db: db}
}

func (r *PortfolioRepository) SaveAll(items []domain.PortfolioItem) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&domain.PortfolioItem{}).Error; err != nil {
			return err
		}
		return tx.Create(&items).Error
	})
}

func (r *PortfolioRepository) FindAll() ([]domain.PortfolioItem, error) {
	var items []domain.PortfolioItem
	if err := r.db.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}
