package config

import "time"

const (
	DefaultEssaysFile   = "endg-urls"
	DefaultWordBankFile = "words.txt"

	// Processing limits
	MaxConcurrentWorkers = 20
	DefaultTopWordsCount = 10
	ProcessingTimeout    = 1 * time.Minute

	// Rate limiting
	DefaultRateLimit = 100 // requests per second

	// Channel buffer sizes
	EssayStreamBufferSize  = 20
	ErrorChannelBufferSize = 100

	// HTTP fetching
	MaxHTTPWorkers = 200
	HTTPRetryDelay = 200 * time.Millisecond

	// Progress reporting
	ProgressReportInterval = 10 // report every N essays

	// Validation limits
	MinWorkers       = 1
	MaxWorkers       = 1000
	MinTopWordsCount = 1
	MaxTopWordsCount = 1000
	MinRateLimit     = 1
	MaxRateLimit     = 1000
	MinBufferSize    = 1
	MaxBufferSize    = 10000
	MinTimeout       = 1 * time.Second
	MaxTimeout       = 1 * time.Hour
)
