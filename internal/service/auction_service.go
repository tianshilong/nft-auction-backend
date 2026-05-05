package service

import (
	"nft-auction-backend/internal/database"
	"nft-auction-backend/internal/repository"
)

type AuctionService struct {
	auctionRepo *repository.AuctionRepo
	bidRepo     *repository.BidRepo
}

func NewAuctionService(ar *repository.AuctionRepo, br *repository.BidRepo) *AuctionService {
	return &AuctionService{auctionRepo: ar, bidRepo: br}
}

func (s *AuctionService) ListAuctions(filters repository.AuctionFilters) ([]database.Auction, int64, error) {
	return s.auctionRepo.List(filters)
}

func (s *AuctionService) GetBidsByAuction(auctionID uint64) ([]database.Bid, error) {
	return s.bidRepo.GetByAuction(auctionID)
}
