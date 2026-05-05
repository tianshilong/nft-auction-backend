package main

import (
	"log"
	"nft-auction-backend/configs"
	"nft-auction-backend/internal/database"
	"nft-auction-backend/internal/ethclient"
	"nft-auction-backend/internal/handler"
	"nft-auction-backend/internal/repository"
	"nft-auction-backend/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg, err := configs.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Auction address: %s, NFT address: %s", cfg.AuctionAddress, cfg.MyNFTAddress)

	// 初始化数据库
	db, err := database.InitDB(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to init DB: %v", err)
	}

	// 初始化事件轮询器
	poller, err := ethclient.NewEventPoller(cfg, db)
	if err != nil {
		log.Fatalf("Failed to create event poller: %v", err)
	}
	go poller.Start()
	defer poller.Stop()

	// 初始化数据仓库
	auctionRepo := repository.NewAuctionRepo(db)
	bidRepo := repository.NewBidRepo(db)
	nftRepo := repository.NewNftRepo(db)

	// 初始化服务
	auctionSvc := service.NewAuctionService(auctionRepo, bidRepo)
	nftSvc := service.NewNFTService(nftRepo)
	statsSvc := service.NewStatsService(auctionRepo, bidRepo)

	// 初始化HTTP处理器
	auctionH := handler.NewAuctionHandler(auctionSvc)
	nftH := handler.NewNFTHandler(nftSvc)
	statsH := handler.NewStatsHandler(statsSvc)

	// 配置路由
	router := gin.Default()
	api := router.Group("/api")
	{
		api.GET("/auctions", auctionH.List)
		api.GET("/auctions/:id/bids", auctionH.Bids)
		api.GET("/nfts/:address", nftH.ListByOwner)
		api.GET("/stats", statsH.GetStats)
	}

	// 启动HTTP服务
	log.Printf("Server starting on %s", cfg.ServerPort)
	if err := router.Run(cfg.ServerPort); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
