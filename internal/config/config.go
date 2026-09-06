package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Env        string
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string
	ServerPort string
}

func Load() (*Config, error) {

	err := godotenv.Load()

	if err != nil {
		return nil, err
	}

	return &Config{
		Env:        os.Getenv("ENV"),
		DBPort:     os.Getenv("DB_PORT"),
		DBName:     os.Getenv("DB_NAME"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		ServerPort: os.Getenv("SERVER_PORT"),
	}, nil
}
