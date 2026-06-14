package app

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Addr                 string
	BaseURL              string
	AppName              string
	Secret               string
	PostgresDSN          string
	BotToken             string
	CompanyLat           string
	CompanyLon           string
	CompanyName          string
	CompanyAddress       string
	DefaultAdminEmail    string
	DefaultAdminPassword string
	DeveloperEmail       string
}

func LoadConfig() Config {
	_ = godotenv.Load()
	cfg := Config{
		Addr:                 env("APP_ADDR", ":8080"),
		BaseURL:              env("APP_BASE_URL", "http://localhost:8080"),
		AppName:              env("APP_NAME", "Partner Sewing Ecosystem"),
		Secret:               env("APP_SECRET", "change-me"),
		PostgresDSN:          env("POSTGRES_DSN", "postgres://postgres:73237323Qwa@localhost:5432/Sewing?sslmode=disable"),
		BotToken:             env("BOT_TOKEN", ""),
		CompanyLat:           env("COMPANY_LAT", "56.129057"),
		CompanyLon:           env("COMPANY_LON", "47.251026"),
		CompanyName:          env("COMPANY_NAME", "Партнёр"),
		CompanyAddress:       env("COMPANY_ADDRESS", "г. Фурманов, ул. Социалистический Посёлок 4"),
		DefaultAdminEmail:    env("DEFAULT_ADMIN_EMAIL", "adimn@gmail.com"),
		DefaultAdminPassword: env("DEFAULT_ADMIN_PASSWORD", "123456"),
		DeveloperEmail:       env("DEVELOPER_EMAIL", "kapustser@gmail.com"),
	}
	if cfg.Secret == "change-me" {
		log.Println("warning: APP_SECRET uses default value")
	}
	return cfg
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
