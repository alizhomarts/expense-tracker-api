package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Config struct {
	AppPort          string `env:"APP_PORT" env-default:"8989"`
	DBHost           string `env:"DB_HOST" env-default:"localhost"`
	DBPort           string `env:"DB_PORT" env-default:"5432"`
	DBUser           string `env:"DB_USER" env-default:"postgres"`
	DBPassword       string `env:"DB_PASSWORD" env-default:"postgres"`
	DBName           string `env:"DB_NAME" env-default:"expense_tracker"`
	JWTAccessSecret  string `env:"JWT_ACCESS_SECRET" env-required:"true"`
	JWTRefreshSecret string `env:"JWT_REFRESH_SECRET" env-required:"true"`

	OpenAIAPIKey string `env:"OPENAI_API_KEY" env-required:"true"`
	OpenAIModel  string `env:"OPENAI_MODEL" env-default:"gpt-5.3-chat-latest"`
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println(".env file not found, using system env")
	}

	return &Config{
		AppPort:          os.Getenv("APP_PORT"),
		DBHost:           os.Getenv("DB_HOST"),
		DBPort:           os.Getenv("DB_PORT"),
		DBUser:           os.Getenv("DB_USER"),
		DBPassword:       os.Getenv("DB_PASSWORD"),
		DBName:           os.Getenv("DB_NAME"),
		JWTAccessSecret:  os.Getenv("JWT_ACCESS_SECRET"),
		JWTRefreshSecret: os.Getenv("JWT_REFRESH_SECRET"),
		OpenAIAPIKey:     os.Getenv("OPENAI_API_KEY"),
		OpenAIModel:      os.Getenv("OPENAI_MODEL"),
	}
}
