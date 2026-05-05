package eventhandler

import (
	_ "embed" // 必须导入，否则无法使用 go:embed
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"nft-auction-backend/internal/database"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"gorm.io/gorm"
)

//go:embed abis/NFTMarketAuctionV2.abi.json
var auctionABIBytes []byte

//go:embed abis/MyNFT.abi.json
var nftABIBytes []byte

type Handler struct {
	db         *gorm.DB // 数据库句柄，用于 CUID 操作
	auctionABI abi.ABI  // 拍卖合约的 ABI 解析器
	nftABI     abi.ABI  // NFT 合约的 ABI 解析器
}

func NewHandler(db *gorm.DB) (*Handler, error) {
	// auction ABI
	// auctionAbiBytes, err := os.ReadFile("../../contracts/NFTMarketAuctionV2.abi.json")
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to read auction ABI: %w", err)
	// }
	// auctionParsed, err := abi.JSON(strings.NewReader(string(auctionAbiBytes)))
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to parse auction ABI: %w", err)
	// }

	// // NFT ABI
	// nftAbiBytes, err := os.ReadFile("../../contracts/MyNFT.abi.json")
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to read NFT ABI: %w", err)
	// }
	// nftParsed, err := abi.JSON(strings.NewReader(string(nftAbiBytes)))
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to parse NFT ABI: %w", err)
	// }

	auctionParsed, err := abi.JSON(strings.NewReader(string(auctionABIBytes)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse auction ABI: %w", err)
	}

	nftParsed, err := abi.JSON(strings.NewReader(string(nftABIBytes)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse NFT ABI: %w", err)
	}

	return &Handler{
		db:         db,
		auctionABI: auctionParsed,
		nftABI:     nftParsed,
	}, nil
}

func (h *Handler) Process(vLog types.Log) error {
	log.Printf("[DEBUG] --------------------Process log: addr=%s tx=%s topics[0]=%x", vLog.Address.Hex(), vLog.TxHash.Hex(), vLog.Topics[0])
	// 尝试将日志的第一个 Topic 与拍卖合约的 ABI 中的事件签名匹配
	event, err := h.auctionABI.EventByID(vLog.Topics[0])
	if err == nil {
		log.Printf("[DEBUG] Matched auction event: %s", event.Name)
		return h.handleAuctionEvent(event, vLog)
	}

	// 如果拍卖合约未匹配，再尝试与 NFT 合约的 ABI 匹配
	event, err = h.nftABI.EventByID(vLog.Topics[0])
	if err == nil {
		return h.handleNFTEvent(event, vLog)
	}
	return nil
}

// ---------- 事件路由 ----------
func (h *Handler) handleAuctionEvent(event *abi.Event, vLog types.Log) error {
	switch event.Name {
	case "AuctionCreated":
		return h.handleAuctionCreated(vLog)
	case "AuctionStarted":
		return h.handleAuctionStarted(vLog)
	case "BidPlaced":
		return h.handleBidPlaced(vLog)
	case "AuctionEnded":
		return h.handleAuctionEnded(vLog)
	case "AuctionCancelled":
		return h.handleAuctionCancelled(vLog)
	case "NFTClaimed":
		return h.handleNFTClaimed(vLog)
	case "FeeExtracted":
		return h.handleFeeExtracted(vLog)
	}
	return nil
}

func (h *Handler) handleNFTEvent(event *abi.Event, vLog types.Log) error {
	if event.Name == "Transfer" {
		return h.handleTransfer(vLog)
	}
	return nil
}

// ---------- 具体事件处理 ----------
func (h *Handler) handleAuctionCreated(vLog types.Log) error {
	log.Println("[DEBUG] --------------------handleAuctionCreated called")
	if len(vLog.Topics) < 4 {
		return nil
	}

	// 1. 解析 indexed 参数：直接从 Topics 中读取
	// Topics[1] 是拍卖ID的uint256哈希（32字节）
	auctionId := new(big.Int).SetBytes(vLog.Topics[1].Bytes()).Uint64()
	// Topics[2], Topics[3] 是地址类型，BytesToAddress 会自动取后20字节
	seller := common.BytesToAddress(vLog.Topics[2].Bytes())
	tokenId := new(big.Int).SetBytes(vLog.Topics[3].Bytes())

	// 2. 解析 non-indexed 参数：从 Data 字段中解析
	var params struct {
		StartPrice *big.Int
		Duration   *big.Int
		EndTime    *big.Int
		State      uint8
	}
	if err := h.auctionABI.UnpackIntoInterface(&params, "AuctionCreated", vLog.Data); err != nil {
		return err
	}

	auction := database.Auction{
		AuctionID:   auctionId,
		Seller:      seller.Hex(),
		NftContract: "",
		TokenId:     tokenId.String(),
		StartPrice:  params.StartPrice.String(),
		Duration:    params.Duration.Uint64(),
		EndTime:     time.Unix(params.EndTime.Int64(), 0),
		State:       params.State,
		CreatedAt:   time.Now(),
	}

	if err := h.db.Save(&auction).Error; err != nil {
		// 如果是唯一约束冲突（例如 auction_id 已存在），只打印警告，不阻断流程
		if strings.Contains(err.Error(), "UNIQUE constraint") {
			log.Printf("[WARN] duplicate auction creation ignored (auctionId=%d): %v", auctionId, err)
			return nil
		}
		log.Printf("[ERROR] save auction failed (auctionId=%d): %v", auctionId, err)
		return err
	}
	return nil
}

func (h *Handler) handleAuctionStarted(vLog types.Log) error {
	log.Println("[DEBUG] --------------------handleAuctionStarted called")
	if len(vLog.Topics) < 2 {
		return nil
	}
	auctionId := new(big.Int).SetBytes(vLog.Topics[1].Bytes()).Uint64()
	var params struct {
		EndTime *big.Int
	}
	if err := h.auctionABI.UnpackIntoInterface(&params, "AuctionStarted", vLog.Data); err != nil {
		return err
	}
	return h.db.Model(&database.Auction{}).Where("auction_id = ?", auctionId).
		Updates(map[string]interface{}{
			"state":    uint8(1),
			"end_time": time.Unix(params.EndTime.Int64(), 0),
		}).Error
}

func (h *Handler) handleBidPlaced(vLog types.Log) error {
	log.Println("[DEBUG] handleBidPlaced called")
	if len(vLog.Topics) < 3 {
		return nil
	}
	auctionId := new(big.Int).SetBytes(vLog.Topics[1].Bytes()).Uint64()
	bidder := common.BytesToAddress(vLog.Topics[2].Bytes())
	var params struct {
		Amount   *big.Int
		UsdValue *big.Int
	}
	if err := h.auctionABI.UnpackIntoInterface(&params, "BidPlaced", vLog.Data); err != nil {
		return err
	}

	bid := database.Bid{
		AuctionID:   auctionId,
		Bidder:      bidder.Hex(),
		Amount:      params.Amount.String(),
		UsdValue:    params.UsdValue.String(),
		BlockNumber: vLog.BlockNumber,
		TxHash:      vLog.TxHash.Hex(),
		CreatedAt:   time.Now(),
	}
	// 忽略重复插入
	if err := h.db.Create(&bid).Error; err != nil {
		return nil
	}

	return h.db.Model(&database.Auction{}).Where("auction_id = ?", auctionId).
		Updates(map[string]interface{}{
			"highest_bidder":  bidder.Hex(),
			"highest_bid":     params.Amount.String(),
			"highest_bid_usd": params.UsdValue.String(),
		}).Error
}

func (h *Handler) handleAuctionEnded(vLog types.Log) error {
	log.Println("[DEBUG] handleAuctionEnded called")
	if len(vLog.Topics) < 3 {
		return nil
	}
	auctionId := new(big.Int).SetBytes(vLog.Topics[1].Bytes()).Uint64()
	var params struct {
		State      uint8
		WinningBid *big.Int
	}
	if err := h.auctionABI.UnpackIntoInterface(&params, "AuctionEnded", vLog.Data); err != nil {
		return err
	}
	return h.db.Model(&database.Auction{}).Where("auction_id = ?", auctionId).
		Update("state", params.State).Error
}

func (h *Handler) handleAuctionCancelled(vLog types.Log) error {
	log.Println("[DEBUG] handleAuctionCancelled called")
	if len(vLog.Topics) < 2 {
		return nil
	}
	auctionId := new(big.Int).SetBytes(vLog.Topics[1].Bytes()).Uint64()
	return h.db.Model(&database.Auction{}).Where("auction_id = ?", auctionId).
		Update("state", uint8(4)).Error
}

func (h *Handler) handleNFTClaimed(vLog types.Log) error {
	log.Println("[DEBUG] handleNFTClaimed called")
	if len(vLog.Topics) < 3 {
		return nil
	}
	auctionId := new(big.Int).SetBytes(vLog.Topics[1].Bytes()).Uint64()
	return h.db.Model(&database.Auction{}).Where("auction_id = ?", auctionId).
		Update("nft_claimed", true).Error
}

func (h *Handler) handleFeeExtracted(vLog types.Log) error {
	// 暂不处理手续费事件
	return nil
}

func (h *Handler) handleTransfer(vLog types.Log) error {
	log.Println("[DEBUG] handleTransfer called")
	if len(vLog.Topics) != 4 {
		return nil
	}
	from := common.BytesToAddress(vLog.Topics[1].Bytes())
	to := common.BytesToAddress(vLog.Topics[2].Bytes())
	tokenId := new(big.Int).SetBytes(vLog.Topics[3].Bytes()).String()

	zeroAddr := common.Address{}
	if from != zeroAddr {
		h.db.Where("contract = ? AND token_id = ? AND owner = ?",
			vLog.Address.Hex(), tokenId, from.Hex()).Delete(&database.NftToken{})
	}
	if to != zeroAddr {
		var nft database.NftToken
		h.db.Where("contract = ? AND token_id = ?", vLog.Address.Hex(), tokenId).
			FirstOrCreate(&nft, database.NftToken{
				Contract: vLog.Address.Hex(),
				TokenID:  tokenId,
			})
		nft.Owner = to.Hex()
		nft.UpdatedAt = time.Now()
		h.db.Save(&nft)
	}
	return nil
}
