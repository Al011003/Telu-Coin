package main

import (
	"log"
	"telkom_coin_back_end/app"
	"telkom_coin_back_end/config"
)

func main() {
	// Load environment variables
	config.LoadEnv()

	// Initialize database
	config.InitDB()

	// Init app (router, handlers, middleware)
	application := app.NewApp()
	r := application.Router

	// Get port from config
	port := config.AppConfig.Port

	log.Printf("Server starting on port %s", port)
	log.Printf("Database: %s@%s:%s/%s", config.AppConfig.DBUser, config.AppConfig.DBHost, config.AppConfig.DBPort, config.AppConfig.DBName)

	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to run server:", err)
	}
}
