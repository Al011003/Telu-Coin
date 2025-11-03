package web3

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Web3Client handles blockchain interactions
type Web3Client struct {
	client          *ethclient.Client
	contractAddress common.Address
	privateKey      *ecdsa.PrivateKey
	chainID         *big.Int
}

// NewWeb3Client creates a new web3 client
func NewWeb3Client() (*Web3Client, error) {
	// Connect to Ganache
	rpcURL := os.Getenv("BLOCKCHAIN_RPC_URL")
	if rpcURL == "" {
		rpcURL = "http://127.0.0.1:8545"
	}

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, err
	}

	// Get chain ID
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return nil, err
	}

	// Load contract address
	contractAddr := os.Getenv("CONTRACT_ADDRESS")
	if contractAddr == "" {
		log.Fatal("CONTRACT_ADDRESS must be set in environment")
	}

	// Load private key for transactions
	privateKeyHex := os.Getenv("ADMIN_PRIVATE_KEY")
	if privateKeyHex == "" {
		log.Fatal("ADMIN_PRIVATE_KEY must be set in environment")
	}

	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, err
	}

	return &Web3Client{
		client:          client,
		contractAddress: common.HexToAddress(contractAddr),
		privateKey:      privateKey,
		chainID:         chainID,
	}, nil
}

// GetTransactor creates a transactor for contract interactions
func (w *Web3Client) GetTransactor() (*bind.TransactOpts, error) {
	auth, err := bind.NewKeyedTransactorWithChainID(w.privateKey, w.chainID)
	if err != nil {
		return nil, err
	}

	// Get nonce
	nonce, err := w.client.PendingNonceAt(context.Background(), auth.From)
	if err != nil {
		return nil, err
	}

	// Get gas price
	gasPrice, err := w.client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	auth.GasLimit = uint64(300000) // Default gas limit
	auth.GasPrice = gasPrice

	return auth, nil
}

// GetCallOpts creates call options for read-only contract calls
func (w *Web3Client) GetCallOpts() *bind.CallOpts {
	return &bind.CallOpts{
		Pending: true,
		Context: context.Background(),
	}
}

// GetClient returns the ethereum client
func (w *Web3Client) GetClient() *ethclient.Client {
	return w.client
}

// GetContractAddress returns the contract address
func (w *Web3Client) GetContractAddress() common.Address {
	return w.contractAddress
}

// Close closes the client connection
func (w *Web3Client) Close() {
	w.client.Close()
}

// GetTransactorFromPrivateKey creates transactor dari private key
func (wc *Web3Client) GetTransactorFromPrivateKey(privateKeyStr string) (*bind.TransactOpts, error) {
	fmt.Printf("[DEBUG] Creating transactor from private key\n")

	privateKey, err := crypto.HexToECDSA(privateKeyStr)
	if err != nil {
		fmt.Printf("[ERROR] HexToECDSA failed: %v\n", err)
		return nil, err
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(1337))
	if err != nil {
		fmt.Printf("[ERROR] Failed to create transactor: %v\n", err)
		return nil, err
	}
	// Set gas price dan limit
	auth.GasPrice = big.NewInt(1000000000) // 1 Gwei
	auth.GasLimit = uint64(300000)

	fmt.Printf("[DEBUG] Transactor created for: %s\n", auth.From.Hex())
	return auth, nil
}

var web3ClientInstance *Web3Client

// GetWeb3ClientInstance returns singleton instance
func GetWeb3ClientInstance() *Web3Client {
	if web3ClientInstance == nil {
		web3ClientInstance, _ = NewWeb3Client()
	}
	return web3ClientInstance
}
