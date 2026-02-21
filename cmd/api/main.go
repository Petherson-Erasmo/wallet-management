package main

import (
	"log"
	"os"

	"github.com/Petherson-Erasmo/wallet-management/internal/handler"
	"github.com/Petherson-Erasmo/wallet-management/internal/repository"
	"github.com/Petherson-Erasmo/wallet-management/internal/service"
)

func main() {
	dsn := getEnv("DATABASE_DSN", "data/wallet.db")

	if err := os.MkdirAll("data", 0755); err != nil {
		log.Fatalf("falha ao criar diretorio de dados: %v", err)
	}

	db, err := repository.NewDB(dsn)
	if err != nil {
		log.Fatalf("falha ao conectar ao banco de dados: %v", err)
	}
	log.Printf("banco de dados inicializado em: %s", dsn)

	portfolioRepo := repository.NewPortfolioRepository(db)
	recommendationRepo := repository.NewRecommendationRepository(db)

	portfolioSvc := service.NewPortfolioService(portfolioRepo)
	recommendationSvc := service.NewRecommendationService(recommendationRepo)
	contributionSvc := service.NewContributionService(portfolioRepo, recommendationRepo)

	portfolioHandler := handler.NewPortfolioHandler(portfolioSvc)
	recommendationHandler := handler.NewRecommendationHandler(recommendationSvc)
	contributionHandler := handler.NewContributionHandler(contributionSvc)

	router := handler.NewRouter(portfolioHandler, recommendationHandler, contributionHandler)

	port := getEnv("PORT", "8080")
	log.Printf("servidor iniciado na porta %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("falha ao iniciar servidor: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}
