package handler

import (
	"net/http"

	"nft-auction-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type NFTHandler struct {
	svc *service.NFTService
}

func NewNFTHandler(svc *service.NFTService) *NFTHandler {
	return &NFTHandler{svc: svc}
}

func (h *NFTHandler) ListByOwner(c *gin.Context) {
	owner := c.Param("address")
	if owner == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "address required"})
		return
	}
	tokens, err := h.svc.GetNFTsByOwner(owner)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tokens)
}
