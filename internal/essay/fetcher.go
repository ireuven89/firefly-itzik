package essay

import (
	"bufio"
	"context"
	"fmt"
	"github.com/ireuven89/firefly-itzik/internal/models"
	"github.com/ireuven89/firefly-itzik/internal/rateLimiter"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	maxWorkers = 200
)

type EssayFetcher interface {
	StreamEssays(ctx context.Context, essayStream chan<- models.Essay, errorChan chan<- error) error
}

type essayFetcher struct {
	client      *http.Client
	rateLimiter rateLimiter.RateLimiter
	filePath    string
}

func NewEssayFetcher(rateLimiter rateLimiter.RateLimiter, filePath string) EssayFetcher {
	transport := &http.Transport{
		MaxIdleConns:          500,
		MaxIdleConnsPerHost:   100,
		IdleConnTimeout:       30 * time.Second,
		DisableKeepAlives:     false,
		DisableCompression:    false,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
	}

	return &essayFetcher{
		client: &http.Client{
			Timeout:   10 * time.Second,
			Transport: transport,
		},
		rateLimiter: rateLimiter,
		filePath:    filePath,
	}
}

func (ef *essayFetcher) StreamEssays(ctx context.Context, essayStream chan<- models.Essay, errorChan chan<- error) error {
	// Read essay URLs from file (this is fine to load all URLs)
	essayURLs, err := ef.readEssayListFromFile()
	if err != nil {
		return fmt.Errorf("failed to read essay list: %w", err)
	}

	// Create channel for workers
	urlChan := make(chan string, maxWorkers)

	// Start workers
	var wg sync.WaitGroup
	numWorkers := maxWorkers
	if len(essayURLs) < maxWorkers {
		numWorkers = len(essayURLs)
	}

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go ef.essayWorker(ctx, urlChan, essayStream, errorChan, &wg)
	}

	// Send URLs to workers
	go func() {
		defer close(urlChan)
		for _, url := range essayURLs {
			select {
			case urlChan <- url:
			case <-ctx.Done():
				return
			}
		}
	}()

	// Wait for all workers to complete
	wg.Wait()
	return nil
}

func (ef *essayFetcher) essayWorker(ctx context.Context, urlChan <-chan string, resultChan chan<- models.Essay, errorChan chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case url, ok := <-urlChan:
			if !ok {
				return
			}

			essay, err := ef.fetchSingleEssay(ctx, url)
			if err != nil {
				errorChan <- err
				continue
			}

			resultChan <- *essay

		case <-ctx.Done():
			return
		}
	}
}

func (ef *essayFetcher) fetchSingleEssay(ctx context.Context, url string) (*models.Essay, error) {
	if err := ef.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	//retry mechanism
	maxRetries := 3
	var resp *http.Response
	var body []byte
	for attempt := 0; attempt < maxRetries; attempt++ {
		resp, err = ef.client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			body, err = io.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("failed reading %s reponse body %v\n", url, err)
				return nil, fmt.Errorf("failed parsing resp body")
			}
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
		// Don't retry 404s
		if resp != nil && resp.StatusCode == 404 {
			fmt.Printf("essay %s returned status 404 (not found)\n", url)
			return nil, fmt.Errorf("essay %s returned status 404 (not found)", url)
		}
		if attempt < maxRetries-1 {
			time.Sleep(200 * time.Millisecond)
		}
	}

	// Extract title and content from HTML (not JSON)
	title := ef.extractTitle(string(body))
	content := ef.extractTextContent(string(body))

	// Extract text content from HTML if needed
	essay := &models.Essay{
		URL:     url,
		Title:   title,
		Content: content,
	}

	return essay, nil
}

func (ef *essayFetcher) extractTitle(htmlContent string) string {
	titleRegex := regexp.MustCompile(`<title[^>]*>([^<]*)</title>`)
	if matches := titleRegex.FindStringSubmatch(htmlContent); len(matches) > 1 {
		title := strings.TrimSpace(matches[1])
		// Clean common suffixes
		title = strings.TrimSuffix(title, " | Engadget")
		title = strings.TrimSuffix(title, " - Engadget")
		return title
	}

	h1Regex := regexp.MustCompile(`<h1[^>]*>([^<]*)</h1>`)
	if matches := h1Regex.FindStringSubmatch(htmlContent); len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	return "Untitled"
}

func (ef *essayFetcher) extractTextContent(htmlContent string) string {
	// Try to find main article content, excluding headers
	articlePatterns := []string{
		`(?is)<div[^>]*class="[^"]*article-body[^"]*"[^>]*>(.*)</div>\s*(?:<footer|</body)`,
		`(?is)<div[^>]*class="[^"]*post-body[^"]*"[^>]*>(.*)</div>\s*(?:<footer|</body)`,
		`(?is)<div[^>]*class="[^"]*entry-content[^"]*"[^>]*>(.*)</div>\s*(?:<footer|</body)`,
		`(?is)<div[^>]*class="[^"]*content-body[^"]*"[^>]*>(.*)</div>\s*(?:<footer|</body)`,

		// Fallback patterns with greedy matching
		`(?is)<div[^>]*class="[^"]*article-body[^"]*"[^>]*>(.*)</div>`,
		`(?is)<div[^>]*class="[^"]*post-body[^"]*"[^>]*>(.*)</div>`,
		`(?is)<div[^>]*class="[^"]*entry-content[^"]*"[^>]*>(.*)</div>`,
		`(?is)<div[^>]*class="[^"]*content-body[^"]*"[^>]*>(.*)</div>`,

		// Try article tag but skip first heading
		`(?is)<article[^>]*>.*?<h[1-6][^>]*>.*?</h[1-6]>(.*?)</article>`,
	}

	for _, pattern := range articlePatterns {
		regex := regexp.MustCompile(pattern)
		if matches := regex.FindStringSubmatch(htmlContent); len(matches) > 1 {
			content := matches[1]
			return ef.cleanTextContent(content)
		}
	}

	// Fallback: extract paragraphs only (skip headings)
	fmt.Printf("falling back to paragraph extraction")
	return ef.extractParagraphsOnly(htmlContent)
}

func (ef *essayFetcher) extractParagraphsOnly(htmlContent string) string {
	// Extract only paragraph content, skip headings
	paragraphRegex := regexp.MustCompile(`(?is)<p[^>]*>(.*?)</p>`)
	matches := paragraphRegex.FindAllStringSubmatch(htmlContent, -1)

	var paragraphs []string
	for _, match := range matches {
		if len(match) > 1 {
			// Clean each paragraph
			cleaned := ef.cleanTextContent(match[1])
			if len(strings.TrimSpace(cleaned)) > 20 { // Skip very short paragraphs
				paragraphs = append(paragraphs, cleaned)
			}
		}
	}

	return strings.Join(paragraphs, " ")
}

func (ef *essayFetcher) cleanTextContent(content string) string {
	// Remove script and style tags with their content
	scriptRegex := regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
	content = scriptRegex.ReplaceAllString(content, " ")

	styleRegex := regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`)
	content = styleRegex.ReplaceAllString(content, " ")

	// Remove HTML comments
	commentRegex := regexp.MustCompile(`<!--.*?-->`)
	content = commentRegex.ReplaceAllString(content, " ")

	// Remove all remaining HTML tags
	htmlTagRegex := regexp.MustCompile(`<[^>]*>`)
	content = htmlTagRegex.ReplaceAllString(content, " ")

	// Replace HTML entities
	content = strings.ReplaceAll(content, "&amp;", "&")
	content = strings.ReplaceAll(content, "&lt;", "<")
	content = strings.ReplaceAll(content, "&gt;", ">")
	content = strings.ReplaceAll(content, "&quot;", "\"")
	content = strings.ReplaceAll(content, "&#39;", "'")
	content = strings.ReplaceAll(content, "&nbsp;", " ")
	content = strings.ReplaceAll(content, "&mdash;", "—")
	content = strings.ReplaceAll(content, "&ndash;", "–")

	// Clean up excessive whitespace
	spaceRegex := regexp.MustCompile(`\s+`)
	content = spaceRegex.ReplaceAllString(content, " ")

	// Remove common noise patterns
	content = strings.ReplaceAll(content, "Advertisement", "")
	content = strings.ReplaceAll(content, "ADVERTISEMENT", "")

	return strings.TrimSpace(content)
}

func (ef *essayFetcher) readEssayListFromFile() ([]string, error) {
	file, err := os.Open(ef.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open essay list file %s: %w", ef.filePath, err)
	}
	defer file.Close()

	var urls []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip empty lines and comments
		if line != "" && !strings.HasPrefix(line, "#") {
			urls = append(urls, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading essay list file: %w", err)
	}

	fmt.Printf("Found %d essay URLs in %s\n", len(urls), ef.filePath)
	return urls, nil
}
