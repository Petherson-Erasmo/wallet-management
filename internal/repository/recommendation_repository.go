package repository

import (
	"github.com/Petherson-Erasmo/wallet-management/internal/domain"
	"gorm.io/gorm"
)

type RecommendationRepository struct {
	db *gorm.DB
}

func NewRecommendationRepository(db *gorm.DB) *RecommendationRepository {
	return &RecommendationRepository{db: db}
}

func (r *RecommendationRepository) SaveAll(items []domain.RecommendedItem) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&domain.RecommendedItem{}).Error; err != nil {
			return err
		}
		return tx.Create(&items).Error
	})
}

func (r *RecommendationRepository) FindAll() ([]domain.RecommendedItem, error) {
	var items []domain.RecommendedItem
	if err := r.db.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}
