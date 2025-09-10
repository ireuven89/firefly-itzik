package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	// File paths
	EssaysFile   string //`env:"ESSAYS_FILE" default:"essays.txt"`
	WordBankFile string //`env:"WORDBANK_FILE" default:"wordbank.txt"`

	// Processing
	MaxWorkers     int           // `env:"MAX_WORKERS" default:"50"`
	TopWordsCount  int           // `env:"TOP_WORDS_COUNT" default:"20"`
	ProcessTimeout time.Duration // `env:"PROCESS_TIMEOUT" default:"10m"`

	// Rate limiting
	RateLimit int //`env:"RATE_LIMIT" default:"10"`

	// Buffers
	EssayStreamBuffer  int //`env:"ESSAY_STREAM_BUFFER" default:"10"`
	ErrorChannelBuffer int //`env:"ERROR_CHANNEL_BUFFER" default:"100"`
}

func LoadConfig() *Config {
	return &Config{
		EssaysFile:         getEnv("ESSAYS_FILE", DefaultEssaysFile),
		WordBankFile:       getEnv("WORDBANK_FILE", DefaultWordBankFile),
		MaxWorkers:         getEnvAsInt("MAX_WORKERS", MaxConcurrentWorkers),
		TopWordsCount:      getEnvAsInt("TOP_WORDS_COUNT", DefaultTopWordsCount),
		ProcessTimeout:     getEnvAsDuration("PROCESS_TIMEOUT", ProcessingTimeout),
		RateLimit:          getEnvAsInt("RATE_LIMIT", DefaultRateLimit),
		EssayStreamBuffer:  getEnvAsInt("ESSAY_STREAM_BUFFER", EssayStreamBufferSize),
		ErrorChannelBuffer: getEnvAsInt("ERROR_CHANNEL_BUFFER", ErrorChannelBufferSize),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
