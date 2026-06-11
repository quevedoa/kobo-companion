package config

import "os"

type Config struct {
	OpenAIAPIKey  string
	OpenAIModel   string
	DBPath        string
	ListenAddress string
}

func Load() Config {
	return Config{
		OpenAIAPIKey:  os.Getenv("OPENAI_API_KEY"),
		OpenAIModel:   envOrFallback("OPENAI_MODEL", "gpt-5.4-nano"),
		DBPath:        envOrFallback("DATABASE_PATH", "/var/lib/kobo-compantion/jobs.db"),
		ListenAddress: envOrFallback("LISTEN_ADDR", "127.0.0.1:8080"),
	}
}

func envOrFallback(key, fallback string) string {
	envVar, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	return envVar
}
