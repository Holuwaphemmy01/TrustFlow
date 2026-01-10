package chain

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"trustflow/src/internal/config"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type ChainClient struct {
	client     *ethclient.Client
	privateKey *ecdsa.PrivateKey
	address    common.Address
	chainID    *big.Int
}

// NewChainClient initializes the connection and loads the wallet
func NewChainClient(cfg *config.Config) (*ChainClient, error) {
	// 1. Connect to RPC
	client, err := ethclient.Dial(cfg.RPCURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RPC: %w", err)
	}

	// 2. Parse Private Key
	// Strip "0x" prefix if present
	pkStr := strings.TrimPrefix(cfg.PrivateKey, "0x")
	privateKey, err := crypto.HexToECDSA(pkStr)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	// 3. Derive Public Address
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("error casting public key to ECDSA")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// 4. Get Chain ID
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	return &ChainClient{
		client:     client,
		privateKey: privateKey,
		address:    fromAddress,
		chainID:    chainID,
	}, nil
}

// GetBalance returns the balance of the connected wallet in Wei
func (c *ChainClient) GetBalance(ctx context.Context) (*big.Int, error) {
	return c.client.BalanceAt(ctx, c.address, nil)
}

// SuggestGasPrice retrieves the currently suggested gas price
func (c *ChainClient) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	return c.client.SuggestGasPrice(ctx)
}

// EstimateGas tries to estimate the gas needed to execute a specific transaction
func (c *ChainClient) EstimateGas(ctx context.Context, callMsg ethereum.CallMsg) (uint64, error) {
	return c.client.EstimateGas(ctx, callMsg)
}

// GetAddress returns the public address of the wallet
func (c *ChainClient) GetAddress() common.Address {
	return c.address
}

// Close closes the underlying client connection
func (c *ChainClient) Close() {
	c.client.Close()
}
