package repository

import (
	"nft-auction-backend/internal/database"

	"gorm.io/gorm"
)

type NftRepo struct {
	db *gorm.DB
}

func NewNftRepo(db *gorm.DB) *NftRepo {
	return &NftRepo{db: db}
}

func (r *NftRepo) GetByOwner(owner string) ([]database.NftToken, error) {
	var tokens []database.NftToken
	err := r.db.Where("owner = ?", owner).Order("updated_at desc").Find(&tokens).Error
	return tokens, err
}
