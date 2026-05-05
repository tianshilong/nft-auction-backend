package repository

import (
	"nft-auction-backend/internal/database"

	"gorm.io/gorm"
)

type BidRepo struct {
	db *gorm.DB
}

func NewBidRepo(db *gorm.DB) *BidRepo {
	return &BidRepo{db: db}
}

func (r *BidRepo) Count() (int64, error) {
	var count int64
	err := r.db.Model(&database.Bid{}).Count(&count).Error
	return count, err
}

func (r *BidRepo) GetByAuction(auctionID uint64) ([]database.Bid, error) {
	var bids []database.Bid
	err := r.db.Where("auction_id = ?", auctionID).Order("created_at desc").Find(&bids).Error
	return bids, err
}
