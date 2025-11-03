package blockchain

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"telkom_coin_back_end/internal/web3"

	"github.com/ethereum/go-ethereum/common"
)

// BlockchainService handles blockchain operations
type BlockchainService struct {
	NetworkID       string
	web3Client      *web3.Web3Client
	contractService *web3.ContractService
}

// NewBlockchainService creates a new blockchain service
func NewBlockchainService(networkID string) (*BlockchainService, error) {
	// Initialize web3 client
	web3Client, err := web3.NewWeb3Client()
	if err != nil {
		// Fallback to simulation mode if web3 fails
		return &BlockchainService{
			NetworkID:       networkID,
			web3Client:      nil,
			contractService: nil,
		}, nil
	}

	// Initialize contract service
	contractService, err := web3.NewContractService(web3Client)
	if err != nil {
		// Fallback to simulation mode if contract fails
		return &BlockchainService{
			NetworkID:       networkID,
			web3Client:      nil,
			contractService: nil,
		}, nil
	}

	return &BlockchainService{
		NetworkID:       networkID,
		web3Client:      web3Client,
		contractService: contractService,
	}, nil
}

// IsWeb3Enabled checks if web3 is available
func (bs *BlockchainService) IsWeb3Enabled() bool {
	return bs.web3Client != nil && bs.contractService != nil
}

// GetTokenBalance gets token balance from smart contract
func (bs *BlockchainService) GetTokenBalance(address string) (*big.Int, error) {
	if !bs.IsWeb3Enabled() {
		// Fallback to simulation
		return big.NewInt(0), nil
	}

	addr := common.HexToAddress(address)
	return bs.contractService.GetBalance(addr)
}

// RequestTopup creates a top-up request on blockchain (pure blockchain only)
func (bs *BlockchainService) RequestTopup(userAddress string, amount *big.Int, paymentProof string) (string, error) {
	if !bs.IsWeb3Enabled() {
		return "", errors.New("blockchain not available - pure blockchain mode required")
	}

	return bs.contractService.InstantTopup("", amount, paymentProof)
}

// RequestWithdraw creates a withdraw request on blockchain (pure blockchain only)
func (bs *BlockchainService) RequestWithdraw(userAddress string, amount *big.Int, bankAccount string) (string, error) {
	if !bs.IsWeb3Enabled() {
		return "", errors.New("blockchain not available - pure blockchain mode required")
	}

	return bs.contractService.RequestWithdraw("", amount, bankAccount)
}

// TransferTokens transfers tokens between addresses (pure blockchain only)
func (bs *BlockchainService) TransferTokens(fromAddress, toAddress string, amount *big.Int, note string) (string, error) {
	if !bs.IsWeb3Enabled() {
		return "", errors.New("blockchain not available - pure blockchain mode required")
	}

	to := common.HexToAddress(toAddress)
	if note != "" {
		return bs.contractService.PaymentTransfer("", to, amount, note)
	}
	return bs.contractService.Transfer("", to, amount)
}

// Removed simulation functions - pure blockchain mode only

// ValidateTransactionHash validates if a transaction hash is valid
func (bs *BlockchainService) ValidateTransactionHash(hash string) bool {
	// Check if hash starts with 0x and has correct length (66 characters)
	if len(hash) != 66 || hash[:2] != "0x" {
		return false
	}

	// Check if the rest are valid hex characters
	_, err := hex.DecodeString(hash[2:])
	return err == nil
}

// GenerateBlockHash generates a hash for a block of transactions
func (bs *BlockchainService) GenerateBlockHash(txHashes []string, blockNumber int64, previousHash string) string {
	// Combine all transaction hashes
	var combinedHashes string
	for _, hash := range txHashes {
		combinedHashes += hash
	}

	// Create block data
	blockData := fmt.Sprintf("%s:%s:%d:%d",
		previousHash,
		combinedHashes,
		blockNumber,
		time.Now().Unix(),
	)

	// Generate SHA256 hash
	hash := sha256.Sum256([]byte(blockData))
	return "0x" + hex.EncodeToString(hash[:])
}

// GetNextNonce gets real nonce from blockchain
func (bs *BlockchainService) GetNextNonce(address string) (uint64, error) {
	if !bs.IsWeb3Enabled() {
		return 0, errors.New("blockchain not available")
	}

	// Get real nonce from blockchain
	addr := common.HexToAddress(address)
	return bs.web3Client.GetClient().PendingNonceAt(context.Background(), addr)
}

// EstimateGas estimates gas needed for a transaction
func (bs *BlockchainService) EstimateGas(txType string, amount string) int64 {
	// Simple gas estimation based on transaction type
	baseGas := int64(21000) // Base gas for simple transfer

	switch txType {
	case "transfer":
		return baseGas
	case "topup":
		return baseGas + 10000 // Additional gas for topup operations
	case "withdraw":
		return baseGas + 15000 // Additional gas for withdraw operations
	default:
		return baseGas
	}
}

// GetGasPrice returns current gas price
func (bs *BlockchainService) GetGasPrice() int64 {
	// In a real implementation, this would query the network
	// For now, return a fixed gas price (in wei)
	return 20000000000 // 20 Gwei
}

// CalculateTransactionFee calculates the transaction fee
func (bs *BlockchainService) CalculateTransactionFee(txType string, amount string) int64 {
	gasLimit := bs.EstimateGas(txType, amount)
	gasPrice := bs.GetGasPrice()
	return gasLimit * gasPrice
}

// VerifyTransactionSignature verifies transaction signature (placeholder)
func (bs *BlockchainService) VerifyTransactionSignature(txHash, signature, publicKey string) bool {
	// In a real implementation, this would verify the cryptographic signature
	// For now, return true as placeholder
	return len(signature) > 0 && len(publicKey) > 0
}

// BroadcastTransaction validates transaction hash (pure blockchain mode)
func (bs *BlockchainService) BroadcastTransaction(txHash string) error {
	// In pure blockchain mode, transactions are automatically broadcast by Web3
	if !bs.ValidateTransactionHash(txHash) {
		return fmt.Errorf("invalid transaction hash format")
	}
	return nil
}

// GetTransactionStatus gets transaction status from blockchain
func (bs *BlockchainService) GetTransactionStatus(txHash string) (string, error) {
	// In a real implementation, this would query the blockchain
	// For now, simulate different statuses based on hash
	if !bs.ValidateTransactionHash(txHash) {
		return "", fmt.Errorf("invalid transaction hash")
	}

	// Simple simulation: use last character of hash to determine status
	lastChar := txHash[len(txHash)-1:]
	switch lastChar {
	case "0", "1", "2", "3", "4", "5":
		return "confirmed", nil
	case "6", "7", "8":
		return "pending", nil
	default:
		return "failed", nil
	}
}

// GetBlockNumber returns current block number
func (bs *BlockchainService) GetBlockNumber() int64 {
	// In a real implementation, this would query the blockchain
	// For now, simulate block number based on time
	return time.Now().Unix() / 15 // Assuming 15 second block time
}

// ConvertToWei converts amount from token units to wei (smallest unit)
func (bs *BlockchainService) ConvertToWei(amount string) (string, error) {
	// Assuming 18 decimals like Ethereum
	// In a real implementation, this would depend on token decimals
	return amount + "000000000000000000", nil
}

// ConvertFromWei converts amount from wei to token units
func (bs *BlockchainService) ConvertFromWei(amountWei string) (string, error) {
	// Simple conversion assuming 18 decimals
	if len(amountWei) <= 18 {
		return "0", nil
	}

	return amountWei[:len(amountWei)-18], nil
}

// GenerateWalletAddress generates a new wallet address (placeholder)
func (bs *BlockchainService) GenerateWalletAddress() string {
	// In a real implementation, this would generate a proper Ethereum address
	// For now, generate a simple address-like string
	timestamp := strconv.FormatInt(time.Now().UnixNano(), 16)
	hash := sha256.Sum256([]byte(timestamp))
	return "0x" + hex.EncodeToString(hash[:20]) // 20 bytes = 40 hex chars
}

// IsValidAddress checks if an address is valid
func (bs *BlockchainService) IsValidAddress(address string) bool {
	// Check if address starts with 0x and has correct length (42 characters)
	if len(address) != 42 || address[:2] != "0x" {
		return false
	}

	// Check if the rest are valid hex characters
	_, err := hex.DecodeString(address[2:])
	return err == nil
}
