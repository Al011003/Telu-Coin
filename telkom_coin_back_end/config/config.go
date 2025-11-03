package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// AppConfigType holds global app settings
type AppConfigType struct {
	Port      string
	DBUser    string
	DBPass    string
	DBHost    string
	DBPort    string
	DBName    string
	JWTSecret string
}

var AppConfig AppConfigType

// LoadEnv reads .env and sets AppConfig
func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, reading from environment variables")
	}

	AppConfig = AppConfigType{
		Port:      os.Getenv("PORT"),
		DBUser:    os.Getenv("DB_USER"),
		DBPass:    os.Getenv("DB_PASS"),
		DBHost:    os.Getenv("DB_HOST"),
		DBPort:    os.Getenv("DB_PORT"),
		DBName:    os.Getenv("DB_NAME"),
		JWTSecret: os.Getenv("JWT_SECRET"),
	}

	if AppConfig.Port == "" {
		AppConfig.Port = "8080"
	}

	if AppConfig.JWTSecret == "" {
		AppConfig.JWTSecret = "your-secret-key-change-this-in-production"
		log.Println("Warning: Using default JWT secret. Set JWT_SECRET environment variable in production")
	}
}
