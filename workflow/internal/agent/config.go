package agent

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Model     string
	Fallback  string
	MaxTokens int
}

func LoadConfig() Config {

	godotenv.Load()

	return Config{
		Model:     os.Getenv("MODEL_ID"),
		Fallback:  os.Getenv("FALLBACK_MODEL_ID"),
		MaxTokens: 8000,
	}
}
