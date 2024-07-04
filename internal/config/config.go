package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"time-tracker/internal/logger"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
}

func LoadConfig() (*Config, error) {
	logger.Logger.Info("Loading config")
	err := godotenv.Load(".env")
	if err != nil {
		return nil, fmt.Errorf("error loading .env file")
	}

	logger.Logger.Info("Loaded config")
	return &Config{
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
	}, nil
}
