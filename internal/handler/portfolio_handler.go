package handler

import (
	"net/http"

	"github.com/Petherson-Erasmo/wallet-management/internal/service"
	"github.com/gin-gonic/gin"
)

type PortfolioHandler struct {
	svc *service.PortfolioService
}

func NewPortfolioHandler(svc *service.PortfolioService) *PortfolioHandler {
	return &PortfolioHandler{svc: svc}
}

func (h *PortfolioHandler) Import(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "arquivo CSV nao encontrado; envie um campo 'file' no form-data"})
		return
	}
	defer file.Close()

	items, err := h.svc.ImportCSV(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "carteira importada com sucesso",
		"total":   len(items),
		"itens":   items,
	})
}

func (h *PortfolioHandler) GetAll(c *gin.Context) {
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
