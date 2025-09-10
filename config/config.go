package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	// File paths
	EssaysFile   string
	WordBankFile string

	// Processing
	MaxWorkers     int
	TopWordsCount  int
	ProcessTimeout time.Duration

	// Rate limiting
	RateLimit int

	// Buffers
	EssayStreamBuffer  int
	ErrorChannelBuffer int

	// HTTP fetching
	MaxHTTPWorkers int
	HTTPRetryDelay time.Duration
}

func LoadConfig() *Config {
	config := &Config{
		EssaysFile:         getEnv("APP_ESSAYS_FILE", DefaultEssaysFile),
		WordBankFile:       getEnv("APP_WORDBANK_FILE", DefaultWordBankFile),
		MaxWorkers:         getEnvAsInt("APP_MAX_WORKERS", MaxConcurrentWorkers),
		TopWordsCount:      getEnvAsInt("APP_TOP_WORDS_COUNT", DefaultTopWordsCount),
		ProcessTimeout:     getEnvAsDuration("APP_PROCESS_TIMEOUT", ProcessingTimeout),
		RateLimit:          getEnvAsInt("APP_RATE_LIMIT", DefaultRateLimit),
		EssayStreamBuffer:  getEnvAsInt("APP_ESSAY_STREAM_BUFFER", EssayStreamBufferSize),
		ErrorChannelBuffer: getEnvAsInt("APP_ERROR_CHANNEL_BUFFER", ErrorChannelBufferSize),
		MaxHTTPWorkers:     getEnvAsInt("APP_MAX_HTTP_WORKERS", MaxHTTPWorkers),
		HTTPRetryDelay:     getEnvAsDuration("APP_HTTP_RETRY_DELAY", HTTPRetryDelay),
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		panic(fmt.Sprintf("Invalid configuration: %v", err))
	}

	return config
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

// Validate checks if all configuration values are within acceptable ranges
func (c *Config) Validate() error {
	// Validate file paths
	if c.EssaysFile == "" {
		return fmt.Errorf("essays file path cannot be empty")
	}
	if c.WordBankFile == "" {
		return fmt.Errorf("word bank file path cannot be empty")
	}

	// Validate processing limits
	if c.MaxWorkers < MinWorkers || c.MaxWorkers > MaxWorkers {
		return fmt.Errorf("max workers must be between %d and %d, got %d", MinWorkers, MaxWorkers, c.MaxWorkers)
	}
	if c.TopWordsCount < MinTopWordsCount || c.TopWordsCount > MaxTopWordsCount {
		return fmt.Errorf("top words count must be between %d and %d, got %d", MinTopWordsCount, MaxTopWordsCount, c.TopWordsCount)
	}
	if c.ProcessTimeout < MinTimeout || c.ProcessTimeout > MaxTimeout {
		return fmt.Errorf("process timeout must be between %v and %v, got %v", MinTimeout, MaxTimeout, c.ProcessTimeout)
	}

	// Validate rate limiting
	if c.RateLimit < MinRateLimit || c.RateLimit > MaxRateLimit {
		return fmt.Errorf("rate limit must be between %d and %d, got %d", MinRateLimit, MaxRateLimit, c.RateLimit)
	}

	// Validate buffer sizes
	if c.EssayStreamBuffer < MinBufferSize || c.EssayStreamBuffer > MaxBufferSize {
		return fmt.Errorf("essay stream buffer must be between %d and %d, got %d", MinBufferSize, MaxBufferSize, c.EssayStreamBuffer)
	}
	if c.ErrorChannelBuffer < MinBufferSize || c.ErrorChannelBuffer > MaxBufferSize {
		return fmt.Errorf("error channel buffer must be between %d and %d, got %d", MinBufferSize, MaxBufferSize, c.ErrorChannelBuffer)
	}

	// Validate HTTP settings
	if c.MaxHTTPWorkers < MinWorkers || c.MaxHTTPWorkers > MaxWorkers {
		return fmt.Errorf("max HTTP workers must be between %d and %d, got %d", MinWorkers, MaxWorkers, c.MaxHTTPWorkers)
	}
	if c.HTTPRetryDelay < 0 || c.HTTPRetryDelay > 5*time.Second {
		return fmt.Errorf("HTTP retry delay must be between 0 and 5s, got %v", c.HTTPRetryDelay)
	}

	return nil
}
