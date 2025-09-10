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
	EssayStreamBufferSize  = 10
	ErrorChannelBufferSize = 100

	// Progress reporting
	ProgressReportInterval = 10 // report every N essays
)
