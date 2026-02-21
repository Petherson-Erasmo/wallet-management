package handler

import (
	"github.com/gin-gonic/gin"
)

func NewRouter(
	portfolioHandler *PortfolioHandler,
	recommendationHandler *RecommendationHandler,
	contributionHandler *ContributionHandler,
) *gin.Engine {
	router := gin.Default()

	v1 := router.Group("/api/v1")
	{
		v1.POST("/portfolio", portfolioHandler.Import)
		v1.GET("/portfolio", portfolioHandler.GetAll)

		v1.POST("/recommendation", recommendationHandler.Import)
		v1.GET("/recommendation", recommendationHandler.GetAll)

		v1.GET("/contribution", contributionHandler.Calculate)
	}

	return router
}
