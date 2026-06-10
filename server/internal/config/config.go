package config

import "os"

type Config struct {
	OpenAIAPIKey string
	OpenAIModel  string
}

func Load() Config {
	return Config{
		OpenAIAPIKey: os.Getenv("OPENAI_API_KEY"),
		OpenAIModel:  envOrFallback("OPENAI_MODEL", "gpt-5.4-nano"),
	}
}

func envOrFallback(key, fallback string) string {
	envVar, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	return envVar
}
