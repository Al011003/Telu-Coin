package config

import (
	"fmt"
	"log"
	"os"
	"telkom_coin_back_end/internal/models"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() *gorm.DB {
	// Load .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	var db *gorm.DB

	// Try MySQL first
	if os.Getenv("DB_HOST") != "" && os.Getenv("DB_USER") != "" {
		// Build MySQL DSN
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASS"),
			os.Getenv("DB_HOST"),
			os.Getenv("DB_PORT"),
			os.Getenv("DB_NAME"),
		)

		// Try to connect to MySQL
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		if err != nil {
			log.Printf("Failed to connect to MySQL: %v", err)
			log.Println("Falling back to SQLite...")
			db = nil
		} else {
			log.Println("✅ Connected to MySQL database")
		}
	}

	// Fallback to SQLite if MySQL failed or not configured
	if db == nil {
		log.Println("Using SQLite database for development...")
		db, err = gorm.Open(sqlite.Open("telkom_coin.db"), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		if err != nil {
			log.Fatal("Failed to connect to SQLite database: " + err.Error())
		}
		log.Println("✅ Connected to SQLite database")
	}

	// Auto migrate tables
	log.Println("Running auto migration...")
	err = db.AutoMigrate(
		&models.User{},
		&models.Transaction{},
		&models.Balance{},
	)
	if err != nil {
		log.Fatal("Failed to auto migrate: " + err.Error())
	}

	log.Println("✅ Database tables migrated successfully!")

	DB = db
	return db
}

func GetDB() *gorm.DB {
	if DB == nil {
		InitDB() // otomatis init kalau belum ada
	}
	return DB
}
