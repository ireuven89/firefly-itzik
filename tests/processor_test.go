package tests

import (
	"context"
	"github.com/ireuven89/firefly-itzik/internal/wordbank"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestWordBank_LoadFromFile(t *testing.T) {
	// Create temporary test file
	ctx := context.Background()
	testFile := "test_wordbank.txt"
	content := `the 
and 
cat 
dog
# comment line
invalid1
ab
123invalid`

	err := os.WriteFile(testFile, []byte(content), 0644)
	assert.NoError(t, err)
	defer os.Remove(testFile)

	// Test loading
	wb := wordbank.NewWordBank(testFile)
	err = wb.LoadWords(ctx, testFile)
	assert.NoError(t, err)

	// Test valid words
	assert.True(t, wb.Contains("the"))
	assert.True(t, wb.Contains("cat"))
	assert.True(t, wb.Contains("THE")) // Should be case insensitive

	// Test invalid words (too short, non-alphabetic)
	assert.False(t, wb.Contains("ab"))
	assert.False(t, wb.Contains("123invalid"))

	// Test non-existent words
	assert.False(t, wb.Contains("xyz"))
}

func TestWordBank_EmptyFile(t *testing.T) {
	ctx := context.Background()
	testFile := "empty_wordbank.txt"
	err := os.WriteFile(testFile, []byte(""), 0644)
	assert.NoError(t, err)
	defer os.Remove(testFile)

	wb := wordbank.NewWordBank(testFile)
	err = wb.LoadWords(ctx, testFile)
	assert.NoError(t, err)

	assert.False(t, wb.Contains("anything"))
}

func TestWordBank_NonExistentFile(t *testing.T) {
	ctx := context.Background()
	wb := wordbank.NewWordBank("")
	err := wb.LoadWords(ctx, "nonexistent.txt")
	assert.Error(t, err)
}

func TestWordBank_CaseInsensitive(t *testing.T) {
	testFile := "case_test.txt"
	content := "Hello\nWORLD\nTest"

	err := os.WriteFile(testFile, []byte(content), 0644)
	assert.NoError(t, err)
	defer os.Remove(testFile)

	wb := wordbank.NewWordBank(testFile)
	wb.LoadWords(context.Background(), testFile)

	assert.True(t, wb.Contains("hello"))
	assert.True(t, wb.Contains("HELLO"))
	assert.True(t, wb.Contains("Hello"))
	assert.True(t, wb.Contains("world"))
	assert.True(t, wb.Contains("test"))
}
