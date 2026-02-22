package handler

import (
	"errors"
	"net/http"

	"github.com/Petherson-Erasmo/wallet-management/internal/domain"
	"github.com/Petherson-Erasmo/wallet-management/internal/service"
	"github.com/gin-gonic/gin"
)

type RecommendationHandler struct {
	svc *service.RecommendationService
}

func NewRecommendationHandler(svc *service.RecommendationService) *RecommendationHandler {
	return &RecommendationHandler{svc: svc}
}

func (h *RecommendationHandler) Import(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "arquivo CSV nao encontrado; envie um campo 'file' no form-data"})
		return
	}
	defer file.Close()

	items, err := h.svc.ImportCSV(file)
	if err != nil {
		if errors.Is(err, domain.ErrInternal) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "carteira recomendada importada com sucesso",
		"total":   len(items),
		"itens":   items,
	})
}

func (h *RecommendationHandler) GetAll(c *gin.Context) {
	items, err := h.svc.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"total": len(items),
		"itens": items,
	})
}
