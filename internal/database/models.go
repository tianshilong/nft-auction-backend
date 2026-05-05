package database

import "time"

// Auction 拍卖信息
type Auction struct {
	ID             uint   `gorm:"primaryKey"`
	AuctionID      uint64 `gorm:"uniqueIndex;not null"`
	Seller         string `gorm:"index;not null"`
	NftContract    string
	TokenId        string `gorm:"not null"`
	StartPrice     string `gorm:"not null"`
	Duration       uint64
	EndTime        time.Time
	HighestBidder  string
	HighestBid     string
	HighestBidUsd  string
	State          uint8 // 0:Pending, 1:Active, 2:Failed, 3:Successful, 4:Cancelled
	BidToken       string
	Erc20PriceFeed string
	NftClaimed     bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (Auction) TableName() string {
	return "auctions"
}

// Bid 出价记录
type Bid struct {
	ID          uint   `gorm:"primaryKey"`
	AuctionID   uint64 `gorm:"index;not null"`
	Bidder      string `gorm:"index;not null"`
	Amount      string `gorm:"not null"`
	UsdValue    string `gorm:"not null"`
	BlockNumber uint64
	TxHash      string `gorm:"uniqueIndex;not null"`
	CreatedAt   time.Time
}

func (Bid) TableName() string {
	return "bids"
}

// NftToken NFT持有记录
type NftToken struct {
	ID        uint   `gorm:"primaryKey"`
	Contract  string `gorm:"index;not null"`
	TokenID   string `gorm:"not null"`
	Owner     string `gorm:"index;not null"`
	UpdatedAt time.Time
}

func (NftToken) TableName() string {
	return "nft_tokens"
}

// SyncState 同步状态
type SyncState struct {
	ID              uint   `gorm:"primaryKey"`
	Key             string `gorm:"uniqueIndex;not null"`
	LastBlockNumber uint64
}

func (SyncState) TableName() string {
	return "sync_states"
}
