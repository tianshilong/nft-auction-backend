package configs

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	RPCURL          string
	MyNFTAddress    string
	AuctionAddress  string
	DBPath          string
	ServerPort      string
	StartBlock      uint64
	PollIntervalSec int
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load() // 忽略文件不存在的错误

	pollInterval, _ := strconv.Atoi(getEnv("POLL_INTERVAL_SEC", "15"))
	startBlock, _ := strconv.ParseUint(getEnv("START_BLOCK", "0"), 10, 64)

	return &Config{
		RPCURL:          getEnv("RPC_URL", "http://127.0.0.1:8545"),
		MyNFTAddress:    getEnv("MYNFT_ADDRESS", ""),
		AuctionAddress:  getEnv("AUCTION_ADDRESS", ""),
		DBPath:          getEnv("DB_PATH", "data.db"),
		ServerPort:      getEnv("SERVER_PORT", ":8080"),
		StartBlock:      startBlock,
		PollIntervalSec: pollInterval,
	}, nil
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
