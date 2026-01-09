package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	RPCURL     string
	PrivateKey string
}

func LoadConfig() (*Config, error) {
	// Load .env file if it exists (ignore error if not found, e.g. in prod)
	_ = godotenv.Load()

	rpcURL := os.Getenv("RPC_URL")
	if rpcURL == "" {
		return nil, os.ErrNotExist // Simplified error for missing env
	}

	privateKey := os.Getenv("PRIVATE_KEY")
	if privateKey == "" {
		return nil, os.ErrNotExist
	}

	return &Config{
		RPCURL:     rpcURL,
		PrivateKey: privateKey,
	}, nil
}
