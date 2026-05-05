package handler

import (
	"net/http"
	"strconv"

	"nft-auction-backend/internal/repository"
	"nft-auction-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type AuctionHandler struct {
	svc *service.AuctionService
}

func NewAuctionHandler(svc *service.AuctionService) *AuctionHandler {
	return &AuctionHandler{svc: svc}
}

func (h *AuctionHandler) List(c *gin.Context) {
	var state *uint8
	if s := c.Query("state"); s != "" {
		val, err := strconv.ParseUint(s, 10, 8)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid state"})
			return
		}
		st := uint8(val)
		state = &st
	}

	sortBy := c.DefaultQuery("sort", "")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	filters := repository.AuctionFilters{
		State:  state,
		SortBy: sortBy,
		Limit:  limit,
		Offset: offset,
	}

	auctions, total, err := h.svc.ListAuctions(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"total": total,
		"data":  auctions,
	})
}

func (h *AuctionHandler) Bids(c *gin.Context) {
	idStr := c.Param("id")
	auctionID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid auction id"})
		return
	}
	bids, err := h.svc.GetBidsByAuction(auctionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, bids)
}
