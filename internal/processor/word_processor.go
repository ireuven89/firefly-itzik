package processor

import (
	"context"
	"fmt"
	"github.com/ireuven89/firefly-itzik/internal/models"
	"github.com/ireuven89/firefly-itzik/internal/wordbank"
	"regexp"
	"sort"
	"strings"
	"sync"
)

// Configuration constants
const (
	DefaultBatchSize = 100
	BatchWorkers     = 20
	ProgressInterval = 10
)

type WordProcessor interface {
	ProcessEssayStream(ctx context.Context, essayStream <-chan models.Essay, errorChan <-chan error, topN int) ([]models.WordCount, int, int)
}

type wordProcessor struct {
	wordBank  wordbank.WordBank
	wordRegex *regexp.Regexp
}

func NewWordProcessor(wordBank wordbank.WordBank) WordProcessor {
	return &wordProcessor{
		wordBank:  wordBank,
		wordRegex: regexp.MustCompile(`[a-zA-Z]+`),
	}
}

func (wp *wordProcessor) ProcessEssayStream(ctx context.Context, essayStream <-chan models.Essay, errorChan <-chan error, topN int) ([]models.WordCount, int, int) {
	wordCounts := make(map[string]int)
	var totalEssays int
	var totalErrors int

	// Process essays in batches for better performance
	batch := make([]models.Essay, 0, DefaultBatchSize)

	for {
		select {
		case essay, ok := <-essayStream:
			if !ok {
				// Process any remaining essays in the final batch
				totalEssays += wp.processFinalBatch(batch, wordCounts)
				totalErrors += wp.drainErrorsAndCount(errorChan)
				return wp.getTopNWords(wordCounts, topN), totalEssays, totalErrors
			}

			// Collect essays into batch
			batch = append(batch, essay)

			// Process full batch
			if len(batch) >= DefaultBatchSize {
				wp.processBatch(batch, wordCounts)
				totalEssays += len(batch)
				batch = batch[:0] // Reset batch for reuse

				wp.logProgress(totalEssays)
			}

		case err, ok := <-errorChan:
			if !ok {
				errorChan = nil
			} else if err != nil {
				totalErrors++
			}

		case <-ctx.Done():
			return wp.getTopNWords(wordCounts, topN), totalEssays, totalErrors
		}
	}
}

// Process final batch when stream is closed
func (wp *wordProcessor) processFinalBatch(batch []models.Essay, wordCounts map[string]int) int {
	if len(batch) == 0 {
		return 0
	}

	wp.processBatch(batch, wordCounts)
	return len(batch)
}

// Drain remaining errors and count them
func (wp *wordProcessor) drainErrorsAndCount(errorChan <-chan error) int {
	errorCount := 0

	for {
		select {
		case err, ok := <-errorChan:
			if !ok {
				return errorCount
			}
			if err != nil {
				fmt.Printf("Processing error: %v\n", err)
				errorCount++
			}
		default:
			return errorCount
		}
	}
}

// Process a batch of essays concurrently
func (wp *wordProcessor) processBatch(essays []models.Essay, globalWordCounts map[string]int) {
	if len(essays) == 0 {
		return
	}

	// Create channels for worker communication
	essayChan := make(chan models.Essay, len(essays))
	resultChan := make(chan map[string]int, BatchWorkers)

	// Send essays to workers
	for _, essay := range essays {
		essayChan <- essay
	}
	close(essayChan)

	// Start concurrent workers
	var wg sync.WaitGroup
	for i := 0; i < BatchWorkers; i++ {
		wg.Add(1)
		go wp.processEssaysWorker(essayChan, resultChan, &wg)
	}

	// Close results when workers finish
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Merge worker results into global counts
	wp.mergeWorkerResults(resultChan, globalWordCounts)
}

// Worker that processes essays and counts words
func (wp *wordProcessor) processEssaysWorker(essayChan <-chan models.Essay, resultChan chan<- map[string]int, wg *sync.WaitGroup) {
	defer wg.Done()

	localWordCounts := make(map[string]int)

	for essay := range essayChan {
		wp.countWordsInText(essay.Content, localWordCounts)
	}

	resultChan <- localWordCounts
}

// Merge results from all workers
func (wp *wordProcessor) mergeWorkerResults(resultChan <-chan map[string]int, globalWordCounts map[string]int) {
	for workerCounts := range resultChan {
		for word, count := range workerCounts {
			globalWordCounts[word] += count
		}
	}
}

// Log processing progress
func (wp *wordProcessor) logProgress(totalEssays int) {
	if totalEssays%ProgressInterval == 0 {
		fmt.Printf("Processed %d essays...\n", totalEssays)
	}
}

// Count valid words in text
func (wp *wordProcessor) countWordsInText(text string, wordCounts map[string]int) {
	words := wp.wordRegex.FindAllString(text, -1)

	for _, word := range words {
		normalizedWord := strings.ToLower(word)

		if wp.isValidWord(normalizedWord) {
			wordCounts[normalizedWord]++
		}
	}
}

// Check if word meets validation criteria
func (wp *wordProcessor) isValidWord(word string) bool {
	return len(word) >= 3 && wp.wordBank.Contains(word)
}

// Get top N words sorted by count
func (wp *wordProcessor) getTopNWords(wordCounts map[string]int, topN int) []models.WordCount {
	words := make([]models.WordCount, 0, len(wordCounts))

	for word, count := range wordCounts {
		words = append(words, models.WordCount{
			Word:  word,
			Count: count,
		})
	}

	// Sort by count descending, then by word ascending
	sort.Slice(words, func(i, j int) bool {
		if words[i].Count == words[j].Count {
			return words[i].Word < words[j].Word
		}
		return words[i].Count > words[j].Count
	})

	if len(words) < topN {
		return words
	}
	return words[:topN]
}
