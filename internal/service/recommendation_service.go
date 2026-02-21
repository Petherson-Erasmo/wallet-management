package service

import (
	"fmt"
	"io"

	csvparser "github.com/Petherson-Erasmo/wallet-management/internal/csvparser"
	"github.com/Petherson-Erasmo/wallet-management/internal/domain"
	"github.com/Petherson-Erasmo/wallet-management/internal/repository"
)

type RecommendationService struct {
	repo *repository.RecommendationRepository
}

func NewRecommendationService(repo *repository.RecommendationRepository) *RecommendationService {
	return &RecommendationService{repo: repo}
}

func (s *RecommendationService) ImportCSV(r io.Reader) ([]domain.RecommendedItem, error) {
	items, err := csvparser.ParseRecommendationCSV(r)
	if err != nil {
		return nil, fmt.Errorf("erro ao processar CSV da carteira recomendada: %w", err)
	}
	if err := s.repo.SaveAll(items); err != nil {
		return nil, fmt.Errorf("erro ao salvar carteira recomendada: %w", err)
	}
	return items, nil
}

func (s *RecommendationService) GetAll() ([]domain.RecommendedItem, error) {
	return s.repo.FindAll()
}
