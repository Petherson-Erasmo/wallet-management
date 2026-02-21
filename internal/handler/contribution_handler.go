package handler

import (
	"net/http"
	"strconv"

	"github.com/Petherson-Erasmo/wallet-management/internal/service"
	"github.com/gin-gonic/gin"
)

type ContributionHandler struct {
	svc *service.ContributionService
}

func NewContributionHandler(svc *service.ContributionService) *ContributionHandler {
	return &ContributionHandler{svc: svc}
}

func (h *ContributionHandler) Calculate(c *gin.Context) {
	valorStr := c.Query("valor")
	if valorStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "parametro 'valor' e obrigatorio (ex: ?valor=1000.00)"})
		return
	}

	valor, err := strconv.ParseFloat(valorStr, 64)
	if err != nil || valor <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "parametro 'valor' deve ser um numero positivo"})
		return
	}

	result, err := h.svc.Calculate(valor)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
