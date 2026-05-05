package ethclient

import (
	"context"
	"log"
	"math/big"
	"time"

	"nft-auction-backend/configs"
	"nft-auction-backend/internal/eventhandler"
	"nft-auction-backend/internal/repository"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"gorm.io/gorm"
)

type EventPoller struct {
	client       *ethclient.Client     // 以太坊节点连接
	db           *gorm.DB              // 数据库，用于状态记录
	auctionAddr  common.Address        // 拍卖合约地址
	nftAddr      common.Address        // NFT合约地址
	handler      *eventhandler.Handler // 事件解析器
	pollInterval time.Duration         // 轮询间隔
	stopCh       chan struct{}         // 停止信号
	syncRepo     *repository.SyncRepo  // 同步区块号的数据库操作
}

func NewEventPoller(cfg *configs.Config, db *gorm.DB) (*EventPoller, error) {
	// 使用带超时的 context 创建 eth 客户端
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cl, err := ethclient.DialContext(ctx, cfg.RPCURL)
	if err != nil {
		return nil, err
	}

	h, err := eventhandler.NewHandler(db)
	if err != nil {
		cl.Close()
		return nil, err
	}

	return &EventPoller{
		client:       cl,
		db:           db,
		auctionAddr:  common.HexToAddress(cfg.AuctionAddress),
		nftAddr:      common.HexToAddress(cfg.MyNFTAddress),
		handler:      h,
		pollInterval: time.Duration(cfg.PollIntervalSec) * time.Second,
		syncRepo:     repository.NewSyncRepo(db),
	}, nil
}

func (p *EventPoller) Start() {
	p.stopCh = make(chan struct{})
	go p.loop()
}

func (p *EventPoller) Stop() {
	close(p.stopCh)
}

func (p *EventPoller) loop() {
	ticker := time.NewTicker(p.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Println("[DEBUG] loop ticker fired, starting sync...")
			p.syncContract(p.auctionAddr, "auction_last_block")
			p.syncContract(p.nftAddr, "nft_last_block")
		case <-p.stopCh:
			return
		}
	}
}

func (p *EventPoller) syncContract(contractAddr common.Address, syncKey string) {
	log.Printf("[DEBUG] syncContract called for %s, address=%s", syncKey, contractAddr.Hex())

	lastBlock, err := p.syncRepo.GetLastBlock(syncKey)
	if err != nil {
		log.Printf("[ERROR] get last block for %s error: %v", syncKey, err)
		return
	}
	log.Printf("[DEBUG] %s last synced block: %d", syncKey, lastBlock)

	latestBlock, err := p.client.BlockNumber(context.Background())
	if err != nil {
		log.Printf("[ERROR] get block number error: %v", err)
		return
	}
	log.Printf("[DEBUG] %s current latest block: %d", syncKey, latestBlock)

	if latestBlock <= lastBlock {
		log.Printf("[DEBUG] %s no new blocks (latest=%d, last=%d), skipping", syncKey, latestBlock, lastBlock)
		return
	}

	fromBlock := new(big.Int).SetUint64(lastBlock + 1)
	toBlock := new(big.Int).SetUint64(latestBlock)
	log.Printf("[DEBUG] %s fetching logs from %d to %d", syncKey, fromBlock.Uint64(), toBlock.Uint64())

	query := ethereum.FilterQuery{
		FromBlock: fromBlock,
		ToBlock:   toBlock,
		Addresses: []common.Address{contractAddr},
	}
	logs, err := p.client.FilterLogs(context.Background(), query)
	if err != nil {
		log.Printf("[ERROR] filter logs error for %s: %v", contractAddr.Hex(), err)
		return
	}
	log.Printf("[DEBUG] %s got %d logs", syncKey, len(logs))

	for i, vLog := range logs {
		log.Printf("[DEBUG] %s processing log #%d: tx=%s block=%d topics[0]=%x",
			syncKey, i, vLog.TxHash.Hex(), vLog.BlockNumber, vLog.Topics[0])
		if err := p.handler.Process(vLog); err != nil {
			log.Printf("[ERROR] process log error: %v", err)
		}
	}

	if err := p.syncRepo.UpdateLastBlock(syncKey, latestBlock); err != nil {
		log.Printf("[ERROR] update last block for %s error: %v", syncKey, err)
	}
	log.Printf("[DEBUG] %s updated last block to %d", syncKey, latestBlock)
}
