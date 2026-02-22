package service

import (
	"errors"
	"fmt"
	"io"

	csvparser "github.com/Petherson-Erasmo/wallet-management/internal/csvparser"
	"github.com/Petherson-Erasmo/wallet-management/internal/domain"
)

type PortfolioService struct {
	repo PortfolioRepository
}

func NewPortfolioService(repo PortfolioRepository) *PortfolioService {
	return &PortfolioService{repo: repo}
}

func (s *PortfolioService) ImportCSV(r io.Reader) ([]domain.PortfolioItem, error) {
	items, err := csvparser.ParsePortfolioCSV(r)
	if err != nil {
		return nil, fmt.Errorf("erro ao processar CSV da carteira: %w", err)
	}
	if err := s.repo.SaveAll(items); err != nil {
		return nil, fmt.Errorf("erro ao salvar carteira: %w", errors.Join(domain.ErrInternal, err))
	}
	return items, nil
}

func (s *PortfolioService) GetAll() ([]domain.PortfolioItem, error) {
	items, err := s.repo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar carteira: %w", errors.Join(domain.ErrInternal, err))
	}
	return items, nil
}
