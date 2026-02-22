package service

import (
	"errors"
	"fmt"
	"io"

	csvparser "github.com/Petherson-Erasmo/wallet-management/internal/csvparser"
	"github.com/Petherson-Erasmo/wallet-management/internal/domain"
)

type RecommendationService struct {
	repo RecommendationRepository
}

func NewRecommendationService(repo RecommendationRepository) *RecommendationService {
	return &RecommendationService{repo: repo}
}

func (s *RecommendationService) ImportCSV(r io.Reader) ([]domain.RecommendedItem, error) {
	items, err := csvparser.ParseRecommendationCSV(r)
	if err != nil {
		return nil, fmt.Errorf("erro ao processar CSV da carteira recomendada: %w", err)
	}
	if err := s.repo.SaveAll(items); err != nil {
		return nil, fmt.Errorf("erro ao salvar carteira recomendada: %w", errors.Join(domain.ErrInternal, err))
	}
	return items, nil
}

func (s *RecommendationService) GetAll() ([]domain.RecommendedItem, error) {
	items, err := s.repo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar carteira recomendada: %w", errors.Join(domain.ErrInternal, err))
	}
	return items, nil
}
