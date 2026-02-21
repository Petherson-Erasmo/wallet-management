package service

import (
	"fmt"
	"io"

	csvparser "github.com/Petherson-Erasmo/wallet-management/internal/csvparser"
	"github.com/Petherson-Erasmo/wallet-management/internal/domain"
	"github.com/Petherson-Erasmo/wallet-management/internal/repository"
)

type PortfolioService struct {
	repo *repository.PortfolioRepository
}

func NewPortfolioService(repo *repository.PortfolioRepository) *PortfolioService {
	return &PortfolioService{repo: repo}
}

func (s *PortfolioService) ImportCSV(r io.Reader) ([]domain.PortfolioItem, error) {
	items, err := csvparser.ParsePortfolioCSV(r)
	if err != nil {
		return nil, fmt.Errorf("erro ao processar CSV da carteira: %w", err)
	}
	if err := s.repo.SaveAll(items); err != nil {
		return nil, fmt.Errorf("erro ao salvar carteira: %w", err)
	}
	return items, nil
}

func (s *PortfolioService) GetAll() ([]domain.PortfolioItem, error) {
	return s.repo.FindAll()
}
