package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// all app's config (env)
type Config struct {
	AppPort     string
	AppEnv      string
	MongoURI    string
	MongoDBName string
	RedisAddr   string
	RedisPass   string
	RabbitMQURL string
}

// this will read the value from .env and return config struct
func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment vairables")
	}

	return &Config{
		AppPort:     getEnv("APP_PORT", "8080"),
		AppEnv:      getEnv("APP_ENV", "development"),
		MongoURI:    getEnv("MONGO_URI", ""),
		MongoDBName: getEnv("MONGO_DB_NAME", "cinema_db"),
		RedisAddr:   getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPass:   getEnv("REDIS_PASSWORD", ""),
		RabbitMQURL: getEnv("RABBITMQ_URL", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}

	return defaultValue
}
