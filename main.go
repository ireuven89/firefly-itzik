package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ireuven89/firefly-itzik/config"
	"github.com/ireuven89/firefly-itzik/internal/essay"
	"github.com/ireuven89/firefly-itzik/internal/models"
	"github.com/ireuven89/firefly-itzik/internal/processor"
	rateLimiter2 "github.com/ireuven89/firefly-itzik/internal/rateLimiter"
	"github.com/ireuven89/firefly-itzik/internal/wordbank"
	"log"
	"time"
)

func main() {
	cfg := config.LoadConfig()
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ProcessTimeout)
	defer cancel()

	// Initialize components
	rateLimiter := rateLimiter2.NewRateLimiter(cfg.RateLimit, time.Second)
	essayFetcher := essay.NewEssayFetcher(rateLimiter, cfg.EssaysFile, cfg.MaxHTTPWorkers, cfg.HTTPRetryDelay)
	wordBank := wordbank.NewWordBank(cfg.WordBankFile)
	wordProcessor := processor.NewWordProcessor(wordBank)
	essayStream := make(chan models.Essay, cfg.EssayStreamBuffer)
	errorChan := make(chan error, cfg.ErrorChannelBuffer)

	// Start streaming essays
	go func() {
		defer close(essayStream)
		defer close(errorChan)

		if err := essayFetcher.StreamEssays(ctx, essayStream, errorChan); err != nil {
			log.Printf("Error streaming essays: %v", err)
		}
	}()

	// Process essays as they stream
	topWords, totalEssays, totalErrors := wordProcessor.ProcessEssayStream(ctx, essayStream, errorChan, cfg.TopWordsCount)

	if totalErrors > 0 {
		log.Printf("Warning: %d errors occurred", totalErrors)
	}

	// Output results
	output := models.Output{
		TopWords:    topWords,
		TotalEssays: totalEssays,
		Timestamp:   time.Now().UTC(),
	}

	jsonOutput, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal output: %v", err)
	}

	fmt.Println(string(jsonOutput))
}
