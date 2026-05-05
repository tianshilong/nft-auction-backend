package service

import (
	"nft-auction-backend/internal/repository"
)

type StatsService struct {
	auctionRepo *repository.AuctionRepo
	bidRepo     *repository.BidRepo
}

func NewStatsService(ar *repository.AuctionRepo, br *repository.BidRepo) *StatsService {
	return &StatsService{auctionRepo: ar, bidRepo: br}
}

func (s *StatsService) GetStats() (map[string]interface{}, error) {
	totalAuctions, err := s.auctionRepo.Count()
	if err != nil {
		return nil, err
	}
	totalBids, err := s.bidRepo.Count()
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"total_auctions": totalAuctions,
		"total_bids":     totalBids,
	}, nil
}
