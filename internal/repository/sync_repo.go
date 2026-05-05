package repository

import (
	"nft-auction-backend/internal/database"

	"gorm.io/gorm"
)

type SyncRepo struct {
	db *gorm.DB
}

func NewSyncRepo(db *gorm.DB) *SyncRepo {
	return &SyncRepo{db: db}
}

func (r *SyncRepo) GetLastBlock(key string) (uint64, error) {
	var state database.SyncState
	err := r.db.Where("key = ?", key).First(&state).Error
	if err != nil {
		return 0, err
	}
	return state.LastBlockNumber, nil
}

func (r *SyncRepo) UpdateLastBlock(key string, blockNum uint64) error {
	return r.db.Model(&database.SyncState{}).Where("key = ?", key).Update("last_block_number", blockNum).Error
}
