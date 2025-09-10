package wordbank

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"
)

type WordBank interface {
	Contains(word string) bool
	LoadWords(ctx context.Context, path string) error
}

type wordBank struct {
	words    map[string]bool
	bankPath string
}

func NewWordBank(path string) WordBank {
	wb := &wordBank{
		words: make(map[string]bool),
	}

	// Load words during initialization
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := wb.LoadWords(ctx, path); err != nil {
		fmt.Printf("Warning: Failed to load word bank: %v\n", err)
	}

	return wb
}

func (wb *wordBank) Contains(word string) bool {
	return wb.words[strings.ToLower(word)]
}

func (wb *wordBank) LoadWords(ctx context.Context, path string) error {
	// Open the word bank file
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open word bank file: %w", err)
	}
	defer file.Close()

	// Read words from file
	var words []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip empty lines and comments
		if line != "" && !strings.HasPrefix(line, "#") {
			words = append(words, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading word bank file: %w", err)
	}

	// Build word map for O(1) lookup (rest stays the same)
	wb.words = make(map[string]bool, len(words))
	for _, word := range words {
		normalizedWord := strings.ToLower(word)
		if len(normalizedWord) >= 3 && isAlphabetic(normalizedWord) {
			wb.words[normalizedWord] = true
		}
	}

	fmt.Printf("Loaded %d valid words from word bank file\n", len(wb.words))

	return nil
}

func isAlphabetic(s string) bool {
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')) {
			return false
		}
	}
	return true
}
