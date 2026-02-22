package service

import "github.com/Petherson-Erasmo/wallet-management/internal/domain"

// PortfolioRepository define o contrato que qualquer repositório de carteira deve satisfazer.
// Isso permite trocar a implementação concreta (SQLite, Postgres, mock) sem alterar os services.
type PortfolioRepository interface {
	FindAll() ([]domain.PortfolioItem, error)
	SaveAll(items []domain.PortfolioItem) error
}

// RecommendationRepository define o contrato que qualquer repositório de recomendações deve satisfazer.
type RecommendationRepository interface {
	FindAll() ([]domain.RecommendedItem, error)
	SaveAll(items []domain.RecommendedItem) error
}
