package web3

import (
	"context"
	"errors"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

// ContractService handles smart contract interactions
type ContractService struct {
	client   *Web3Client
	contract *bind.BoundContract
	abi      abi.ABI
}

// NewContractService creates a new contract service
func NewContractService(client *Web3Client) (*ContractService, error) {
	contractABI, err := abi.JSON(strings.NewReader(PaymentTokenABI))
	if err != nil {
		return nil, err
	}

	// Asumsi dari kode Anda, GetClient() mengembalikan *ethclient.Client
	boundContract := bind.NewBoundContract(client.GetContractAddress(), contractABI, client.GetClient(), client.GetClient(), client.GetClient())

	return &ContractService{
		client:   client,
		contract: boundContract,
		abi:      contractABI,
	}, nil
}

// TopupRequest represents a top-up request
type TopupRequest struct {
	User         common.Address
	Amount       *big.Int
	PaymentProof string
	Timestamp    *big.Int
	Processed    bool
}

// WithdrawRequest represents a withdraw request
type WithdrawRequest struct {
	User        common.Address
	Amount      *big.Int
	BankAccount string
	Timestamp   *big.Int
	Processed   bool
}

func (cs *ContractService) createManualTransactor(privateKeyHex string) (*bind.TransactOpts, error) {
	// 1. Konversi private key
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, errors.New("failed to convert private key: " + err.Error())
	}

	// 2. Ambil Chain ID. Asumsi GetClient() mengembalikan *ethclient.Client
	chainID, err := cs.client.GetClient().ChainID(context.Background())
	if err != nil {
		return nil, errors.New("failed to get chain ID: " + err.Error())
	}

	// 3. Buat transactor dasar
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return nil, errors.New("failed to create transactor: " + err.Error())
	}

	// 4. Atur Nonce
	fromAddress := auth.From
	nonce, err := cs.client.GetClient().PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return nil, errors.New("failed to get nonce: " + err.Error())
	}
	auth.Nonce = big.NewInt(int64(nonce))

	// 5. Atur Gas Price
	gasPrice, err := cs.client.GetClient().SuggestGasPrice(context.Background())
	if err != nil {
		return nil, errors.New("failed to suggest gas price: " + err.Error())
	}
	auth.GasPrice = gasPrice

	// 6. ATUR GAS LIMIT MANUAL (INI KUNCINYA!)
	auth.GasLimit = uint64(300000) // Nilai aman untuk fungsi standar

	// 7. Log semua detail untuk debugging
	log.Println("====================== [ DEBUG: New Transactor Created ] ======================")
	log.Printf("[DEBUG] Signer Address: %s", auth.From.Hex())
	log.Printf("[DEBUG] Chain ID: %s", chainID.String())
	log.Printf("[DEBUG] Nonce: %s", auth.Nonce.String())
	log.Printf("[DEBUG] Gas Price: %s wei", auth.GasPrice.String())
	log.Printf("[DEBUG] Gas Limit (MANUALLY SET): %d", auth.GasLimit)
	log.Println("==============================================================================")

	return auth, nil
}

func (cs *ContractService) InstantTopup(userPrivateKey string, amount *big.Int, paymentProof string) (string, error) {
	// Ganti cs.client.GetTransactor() dengan helper baru kita
	auth, err := cs.createManualTransactor(userPrivateKey)
	if err != nil {
		return "", err
	}

	tx, err := cs.contract.Transact(auth, "instantTopup", amount, paymentProof)
	if err != nil {
		return "", err
	}
	return tx.Hash().Hex(), nil
}

// ProcessTopup processes a pending top-up request
func (cs *ContractService) ProcessTopup(requestId [32]byte) (string, error) {
	// Untuk fungsi admin, kita pakai private key dari environment variable
	adminPrivateKey := os.Getenv("ADMIN_PRIVATE_KEY")
	if adminPrivateKey == "" {
		return "", errors.New("ADMIN_PRIVATE_KEY environment variable not set")
	}

	auth, err := cs.createManualTransactor(adminPrivateKey)
	if err != nil {
		return "", err
	}

	tx, err := cs.contract.Transact(auth, "processTopup", requestId)
	if err != nil {
		return "", err
	}
	return tx.Hash().Hex(), nil
}

// RequestWithdraw creates a withdraw request and burns tokens
func (cs *ContractService) RequestWithdraw(userPrivateKey string, amount *big.Int, bankAccount string) (string, error) {
	auth, err := cs.createManualTransactor(userPrivateKey)
	if err != nil {
		return "", err
	}

	tx, err := cs.contract.Transact(auth, "requestWithdraw", amount, bankAccount)
	if err != nil {
		return "", err
	}
	return tx.Hash().Hex(), nil
}

// PaymentTransfer transfers tokens with a note
func (cs *ContractService) PaymentTransfer(userPrivateKey string, to common.Address, amount *big.Int, note string) (string, error) {
	auth, err := cs.createManualTransactor(userPrivateKey)
	if err != nil {
		return "", err
	}

	tx, err := cs.contract.Transact(auth, "paymentTransfer", to, amount, note)
	if err != nil {
		return "", err
	}
	return tx.Hash().Hex(), nil
}

// Transfer standard ERC20 transfer
func (cs *ContractService) Transfer(userPrivateKey string, to common.Address, amount *big.Int) (string, error) {
	// Sekarang fungsi ini jadi simpel dan konsisten
	auth, err := cs.createManualTransactor(userPrivateKey)
	if err != nil {
		return "", err
	}

	tx, err := cs.contract.Transact(auth, "transfer", to, amount)
	if err != nil {
		return "", err
	}
	return tx.Hash().Hex(), nil
}

// GetBalance gets token balance for an address
func (cs *ContractService) GetBalance(address common.Address) (*big.Int, error) {
	var result []interface{}
	err := cs.contract.Call(cs.client.GetCallOpts(), &result, "balanceOf", address)
	if err != nil {
		return nil, err
	}
	return result[0].(*big.Int), nil
}

// GetTotalSupply gets total token supply
func (cs *ContractService) GetTotalSupply() (*big.Int, error) {
	var result []interface{}
	err := cs.contract.Call(cs.client.GetCallOpts(), &result, "totalSupply")
	if err != nil {
		return nil, err
	}
	return result[0].(*big.Int), nil
}

// ProcessTopup processes a pending top-up request

// GetTopupRequest gets details of a top-up request
func (cs *ContractService) GetTopupRequest(requestId [32]byte) (*TopupRequest, error) {
	var result []interface{}
	err := cs.contract.Call(cs.client.GetCallOpts(), &result, "getTopupRequest", requestId)
	if err != nil {
		return nil, err
	}

	return &TopupRequest{
		User:         result[0].(common.Address),
		Amount:       result[1].(*big.Int),
		PaymentProof: result[2].(string),
		Timestamp:    result[3].(*big.Int),
		Processed:    result[4].(bool),
	}, nil
}

// GetWithdrawRequest gets details of a withdraw request
func (cs *ContractService) GetWithdrawRequest(requestId [32]byte) (*WithdrawRequest, error) {
	var result []interface{}
	err := cs.contract.Call(cs.client.GetCallOpts(), &result, "getWithdrawRequest", requestId)
	if err != nil {
		return nil, err
	}

	return &WithdrawRequest{
		User:        result[0].(common.Address),
		Amount:      result[1].(*big.Int),
		BankAccount: result[2].(string),
		Timestamp:   result[3].(*big.Int),
		Processed:   result[4].(bool),
	}, nil
}

// GetTransactionReceipt gets transaction receipt
func (cs *ContractService) GetTransactionReceipt(txHash string) (*types.Receipt, error) {
	hash := common.HexToHash(txHash)
	return cs.client.GetClient().TransactionReceipt(context.Background(), hash)
}

// GetMinMintAmount gets minimum mint amount
func (cs *ContractService) GetMinMintAmount() (*big.Int, error) {
	var result []interface{}
	err := cs.contract.Call(cs.client.GetCallOpts(), &result, "minMintAmount")
	if err != nil {
		return nil, err
	}
	return result[0].(*big.Int), nil
}

// GetMinBurnAmount gets minimum burn amount
func (cs *ContractService) GetMinBurnAmount() (*big.Int, error) {
	var result []interface{}
	err := cs.contract.Call(cs.client.GetCallOpts(), &result, "minBurnAmount")
	if err != nil {
		return nil, err
	}
	return result[0].(*big.Int), nil
}

// GetExchangeRate gets current exchange rate
func (cs *ContractService) GetExchangeRate() (*big.Int, error) {
	var result []interface{}
	err := cs.contract.Call(cs.client.GetCallOpts(), &result, "exchangeRate")
	if err != nil {
		return nil, err
	}
	return result[0].(*big.Int), nil
}

func (cs *ContractService) WaitForReceipt(txHash string, timeoutSeconds int) (*types.Receipt, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	txHashHex := common.HexToHash(txHash)

	for {
		select {
		case <-ctx.Done():
			return nil, errors.New("timeout waiting for transaction confirmation")

		case <-ticker.C:
			// Ambil receipt dari blockchain
			receipt, err := cs.client.client.TransactionReceipt(ctx, txHashHex)
			if err == nil {
				// Receipt ketemu!
				if receipt.Status == types.ReceiptStatusFailed {
					return nil, errors.New("transaction reverted on blockchain")
				}
				log.Printf("✅ Transaction mined at block %d", receipt.BlockNumber.Uint64())
				return receipt, nil
			}

			// Belum ketemu, tunggu lagi
			log.Printf("⏳ Waiting for tx %s...", txHash[:10])
		}
	}
}
