# Firefly-Itzik

A high-performance Go application that scrapes articles from Engadget, extracts text content, and analyzes word frequencies using a word bank filter.


## Prerequisites

- Go 1.22.9 or later
- Internet connection (for fetching articles)

## Quick Start

1. **Clone and build:**
```bash
git clone <repository-url>
cd firefly-itzik
go mod download
go build -o firefly-itzik .
```

2. **Run with default settings:**
```bash
./firefly-itzik
```

The application will:
- Load ~415K words from `words.txt`
- Read ~40K URLs from `endg-urls`
- Scrape articles concurrently
- Output top 10 most frequent words as JSON

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd firefly-itzik
```

2. Install dependencies:
```bash
go mod download
```

3. Build the application:
```bash
go build -o firefly-itzik .
```

## Usage

### Basic Usage

Run with default settings:
```bash
./firefly-itzik
```

### Configuration via Environment Variables

The application supports extensive configuration through environment variables. All environment variables are prefixed with `APP_` to avoid conflicts with other applications:


#### File Paths
- `APP_ESSAYS_FILE`: Path to file containing article URLs (default: `endg-urls`)
- `APP_WORDBANK_FILE`: Path to word bank file (default: `words.txt`)

#### Processing Configuration
- `APP_MAX_WORKERS`: Maximum concurrent processing workers (default: `20`, range: 1-1000)
- `APP_TOP_WORDS_COUNT`: Number of top words to return (default: `10`, range: 1-1000)
- `APP_PROCESS_TIMEOUT`: Overall processing timeout (default: `1m`, range: 1s-1h)

#### Rate Limiting
- `APP_RATE_LIMIT`: Requests per second limit (default: `100`, range: 1-1000)

#### Buffer Sizes
- `APP_ESSAY_STREAM_BUFFER`: Essay stream channel buffer size (default: `20`, range: 1-10000)
- `APP_ERROR_CHANNEL_BUFFER`: Error channel buffer size (default: `100`, range: 1-10000)

#### HTTP Configuration
- `APP_MAX_HTTP_WORKERS`: Maximum concurrent HTTP workers (default: `200`, range: 1-1000)
- `APP_HTTP_RETRY_DELAY`: Delay between retry attempts (default: `200ms`, range: 0-5s)

### Example Usage with Custom Configuration

```bash
# Run with custom settings
export APP_MAX_HTTP_WORKERS=50
export APP_RATE_LIMIT=25
export APP_TOP_WORDS_COUNT=20
export APP_PROCESS_TIMEOUT=5m
./firefly-itzik
```

Or run with inline environment variables:
```bash
APP_MAX_HTTP_WORKERS=50 APP_RATE_LIMIT=25 APP_TOP_WORDS_COUNT=20 ./firefly-itzik
```

## Input Files

### Article URLs File (`endg-urls`)
A text file containing one URL per line:
```
https://www.engadget.com/2019/08/25/sony-and-yamaha-sc-1-sociable-cart/
https://www.engadget.com/2019/08/24/trump-tries-to-overturn-ruling-stopping-him-from-blocking-twitte/
https://www.engadget.com/2019/08/24/crime-allegation-in-space/
```

### Word Bank File (`words.txt`)
A text file containing valid words, one per line:
```
the
and
technology
computer
software
```

## Output

The application outputs a JSON object containing:
- `top_words`: Array of word count objects sorted by frequency
- `total_essays`: Total number of essays processed
- `timestamp`: Processing completion timestamp

Example output:
```json
{
  "top_words": [
    {
      "word": "technology",
      "count": 1247
    },
    {
      "word": "device",
      "count": 892
    },
    {
      "word": "software",
      "count": 654
    }
  ],
  "total_essays": 1000,
  "timestamp": "2024-01-15T10:30:45Z"
}
```

## Performance Tuning

### For High-Volume Processing
```bash
export APP_MAX_HTTP_WORKERS=500
export APP_RATE_LIMIT=200
export APP_ESSAY_STREAM_BUFFER=100
export APP_ERROR_CHANNEL_BUFFER=500
```

### For Conservative Scraping
```bash
export APP_MAX_HTTP_WORKERS=10
export APP_RATE_LIMIT=5
export APP_HTTP_RETRY_DELAY=1s
```

### For Memory-Constrained Environments
```bash
export APP_MAX_HTTP_WORKERS=50
export APP_ESSAY_STREAM_BUFFER=10
export APP_ERROR_CHANNEL_BUFFER=50
```

## Testing

Run the test suite:
```bash
go test ./...
```

Run tests with verbose output:
```bash
go test -v ./...
```

## Architecture

The application follows a modular architecture:

- **`config/`**: Configuration management and validation
- **`internal/essay/`**: Web scraping and HTML parsing
- **`internal/processor/`**: Text processing and word counting
- **`internal/wordbank/`**: Word bank management
- **`internal/rateLimiter/`**: Rate limiting implementation
- **`internal/models/`**: Data structures

## Error Handling

The application includes comprehensive error handling:
- HTTP request retries with exponential backoff
- Graceful handling of 404 errors
- Context-based cancellation
- Detailed error reporting

### Common Issues

**Build fails with "no such file or directory"**
- Ensure you're in the project root directory
- Run `go mod download` to install dependencies

**Configuration validation errors**
- Check that all environment variables are within valid ranges
- Ensure file paths exist and are accessible

**HTTP timeout errors**
- Increase `PROCESS_TIMEOUT` for large datasets
- Reduce `RATE_LIMIT` if getting rate limited
- Check internet connectivity

**Memory issues with large datasets**
- Reduce `MAX_HTTP_WORKERS` and buffer sizes
- Process data in smaller batches

### Debug Mode

For debugging, you can increase verbosity by modifying the logging in the source code or using Go's built-in debugging tools:

```bash
go run -race .  # Run with race detection
go run -v .     # Run with verbose output
```
