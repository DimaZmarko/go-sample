package config

import (
	"os"
)

type Config struct {
	PostgresHost     string
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	RedisHost        string
	RedisPort        string
}

func NewConfig() *Config {
	return &Config{
		PostgresHost:     os.Getenv("POSTGRES_HOST"),
		PostgresUser:     os.Getenv("POSTGRES_USER"),
		PostgresPassword: os.Getenv("POSTGRES_PASSWORD"),
		PostgresDB:       os.Getenv("POSTGRES_DB"),
		RedisHost:        os.Getenv("REDIS_HOST"),
		RedisPort:        os.Getenv("REDIS_PORT"),
	}
} 