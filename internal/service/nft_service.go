package service

import (
	"nft-auction-backend/internal/database"
	"nft-auction-backend/internal/repository"
)

type NFTService struct {
	nftRepo *repository.NftRepo
}

func NewNFTService(nr *repository.NftRepo) *NFTService {
	return &NFTService{nftRepo: nr}
}

func (s *NFTService) GetNFTsByOwner(owner string) ([]database.NftToken, error) {
	return s.nftRepo.GetByOwner(owner)
}
