package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	RedisHost        string
	RedisPort        string
}

func NewConfig() *Config {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	config := &Config{
		PostgresHost:     os.Getenv("POSTGRES_HOST"),
		PostgresPort:     os.Getenv("POSTGRES_PORT"),
		PostgresUser:     os.Getenv("POSTGRES_USER"),
		PostgresPassword: os.Getenv("POSTGRES_PASSWORD"),
		PostgresDB:       os.Getenv("POSTGRES_DB"),
		RedisHost:        os.Getenv("REDIS_HOST"),
		RedisPort:        os.Getenv("REDIS_PORT"),
	}

	// Validate required environment variables
	if config.PostgresHost == "" {
		log.Fatal("POSTGRES_HOST environment variable is required")
	}
	if config.PostgresPort == "" {
		log.Fatal("POSTGRES_PORT environment variable is required")
	}
	if config.PostgresUser == "" {
		log.Fatal("POSTGRES_USER environment variable is required")
	}
	if config.PostgresPassword == "" {
		log.Fatal("POSTGRES_PASSWORD environment variable is required")
	}
	if config.PostgresDB == "" {
		log.Fatal("POSTGRES_DB environment variable is required")
	}
	if config.RedisHost == "" {
		log.Fatal("REDIS_HOST environment variable is required")
	}
	if config.RedisPort == "" {
		log.Fatal("REDIS_PORT environment variable is required")
	}

	return config
}
