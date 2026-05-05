package repository

import (
	"nft-auction-backend/internal/database"

	"gorm.io/gorm"
)

type AuctionRepo struct {
	db *gorm.DB
}

func NewAuctionRepo(db *gorm.DB) *AuctionRepo {
	return &AuctionRepo{db: db}
}

// SelectFields 用于干净的返回字段（避免直接返回整个结构体时包含ID等内部字段）
type AuctionFilters struct {
	State  *uint8
	SortBy string // "price_asc", "price_desc", "end_asc", "end_desc"
	Limit  int
	Offset int
}

func (r *AuctionRepo) Count() (int64, error) {
	var count int64
	err := r.db.Model(&database.Auction{}).Count(&count).Error
	return count, err
}

func (r *AuctionRepo) List(filters AuctionFilters) ([]database.Auction, int64, error) {
	query := r.db.Model(&database.Auction{})

	if filters.State != nil {
		query = query.Where("state = ?", *filters.State)
	}

	// 排序
	switch filters.SortBy {
	case "price_asc":
		query = query.Order("cast(start_price as REAL) asc")
	case "price_desc":
		query = query.Order("cast(start_price as REAL) desc")
	case "end_asc":
		query = query.Order("end_time asc")
	case "end_desc":
		query = query.Order("end_time desc")
	default:
		query = query.Order("created_at desc")
	}

	var total int64
	query.Count(&total)

	var auctions []database.Auction
	query.Offset(filters.Offset).Limit(filters.Limit).Find(&auctions)
	return auctions, total, query.Error
}

func (r *AuctionRepo) GetByID(auctionID uint64) (*database.Auction, error) {
	var auction database.Auction
	err := r.db.Where("auction_id = ?", auctionID).First(&auction).Error
	return &auction, err
}
