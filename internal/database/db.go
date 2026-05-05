package database

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func InitDB(dbPath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	// 自动迁移表结构
	err = db.AutoMigrate(&Auction{}, &Bid{}, &NftToken{}, &SyncState{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate: %w", err)
	}

	initSyncState(db, "auction_last_block", 0)
	initSyncState(db, "nft_last_block", 0)

	return db, nil
}

func initSyncState(db *gorm.DB, key string, defaultBlock uint64) {
	var s SyncState
	if err := db.Where("key = ?", key).First(&s).Error; err != nil {
		db.Create(&SyncState{Key: key, LastBlockNumber: defaultBlock})
	}
}
