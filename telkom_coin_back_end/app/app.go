package app

import (
	"log"

	"telkom_coin_back_end/config"
	"telkom_coin_back_end/internal/blockchain"
	"telkom_coin_back_end/internal/handler"
	"telkom_coin_back_end/internal/middleware"
	"telkom_coin_back_end/internal/repository"
	service "telkom_coin_back_end/internal/services"

	"github.com/gin-gonic/gin"
)

type App struct {
	Router *gin.Engine
}

func NewApp() *App {
	r := gin.Default()

	// Repositories
	userRepo := repository.NewUserRepository(config.GetDB())
	balanceRepo := repository.NewBalanceRepository(config.GetDB())
	txRepo := repository.NewTransactionRepository(config.GetDB())

	// Blockchain service
	blockchainService, err := blockchain.NewBlockchainService("1337")
	if err != nil {
		log.Printf("Warning: Failed to initialize blockchain service: %v", err)
	}
	if blockchainService.IsWeb3Enabled() {
		log.Println("✅ Web3 blockchain integration enabled")
	} else {
		log.Println("⚠️  Web3 disabled, using simulation mode")
	}

	// Services
	authService := service.NewAuthService(userRepo)
	authService.SetBalanceRepo(balanceRepo) // Set balance repo for creating initial balance
	pinService := service.NewPinService(userRepo)
	userService := service.NewUserService(userRepo, balanceRepo)
	topupService := service.NewTopupService(userRepo, balanceRepo, txRepo, blockchainService)
	withdrawService := service.NewWithdrawService(userRepo, balanceRepo, txRepo, blockchainService)
	blockchainExplorerService, err := service.NewBlockchainExplorerService()
	if err != nil {
		log.Printf("Warning: Blockchain explorer not available: %v", err)
	}

	// TLC Wallet Service (blockchain-based)
	tlcWalletService := service.NewTLCWalletService(
		userRepo,
		balanceRepo,
		txRepo,
		blockchainService,
		blockchainExplorerService,
	)

	// Blockchain Explorer Service
	

	// Handlers
	authHandler := handler.NewAuthHandler(authService)
	pinHandler := handler.NewPinHandler(pinService)
	userHandler := handler.NewUserHandler(userService, authService)
	topupHandler := handler.NewTopupHandler(topupService)
	withdrawHandler := handler.NewWithdrawHandler(withdrawService, userService)

	// TLC Wallet Handler (blockchain-based)
	tlcWalletHandler := handler.NewTLCWalletHandler(tlcWalletService, userService)

	// Blockchain Explorer Handler
	var blockchainExplorerHandler *handler.BlockchainExplorerHandler
	if blockchainExplorerService != nil {
		blockchainExplorerHandler = handler.NewBlockchainExplorerHandler(blockchainExplorerService)
	}

	// Public routes
	r.POST("/register", authHandler.Register)
	r.POST("/login", authHandler.Login)

	// Protected routes
	auth := r.Group("/api")
	auth.Use(middleware.JWTMiddleware())
	{
		// User management
		auth.GET("/profile", userHandler.GetProfile)
		auth.GET("/profile/detail", userHandler.GetDetailProfile)
		auth.PUT("/profile", userHandler.UpdateProfile)
		auth.POST("/change-password", userHandler.ChangePassword)
		auth.POST("/change-pin", userHandler.ChangePin)
		auth.GET("/account-status", userHandler.GetAccountStatus)
		// KYC routes removed - auto-verified system

		// PIN management
		auth.POST("/pin/set", pinHandler.SetPin)
		auth.POST("/pin/verify", pinHandler.VerifyPin)

		// Balance (Unified - only blockchain balance)
		auth.GET("/balance", tlcWalletHandler.GetTLCBalance)

		// Topup (Simplified - Auto processed, no proof needed)
		auth.POST("/topup", topupHandler.RequestTopup)
		auth.GET("/topup/history", topupHandler.GetTopupHistory)

		transferGroup := auth.Group("/transfer")
		{
			transferGroup.POST("/validate", tlcWalletHandler.ValidateTransfer)
			transferGroup.POST("/execute", tlcWalletHandler.TransferTLC)
		}

		// Transfer (Direct blockchain only)

		// Withdraw (Direct blockchain only)
		auth.POST("/withdraw", withdrawHandler.RequestWithdraw)

		// Transactions (Blockchain only)
		auth.GET("/transactions", tlcWalletHandler.GetTransactionHistory)

		// TLC Blockchain Explorer (Public - like Bitcoin explorer)
		if blockchainExplorerHandler != nil {
			explorer := r.Group("/explorer")
			{
				// All transactions (public)
				explorer.GET("/transactions", blockchainExplorerHandler.GetAllTransactions)
				explorer.GET("/transactions/latest", blockchainExplorerHandler.GetLatestTransactions)
				explorer.GET("/transactions/search", blockchainExplorerHandler.SearchTransactions)
				explorer.GET("/transactions/types", blockchainExplorerHandler.GetTransactionTypes)

				// Specific transaction
				explorer.GET("/tx/:hash", blockchainExplorerHandler.GetTransactionByHash)

				// Address transactions
				explorer.GET("/address/:address", blockchainExplorerHandler.GetTransactionsByAddress)

				// Blockchain stats
				explorer.GET("/stats", blockchainExplorerHandler.GetBlockchainStats)
			}
		}
	}

	// Admin routes removed - no admin system needed for pure blockchain

	return &App{Router: r}
}
