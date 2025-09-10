package essay

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExtractTextContent_ArticleBodyPattern(t *testing.T) {
	fetcher := &essayFetcher{}

	htmlContent := `
    <html>
    <head><title>Test Article</title></head>
    <body>
        <header>Site Header</header>
        <nav>Navigation</nav>
        <div class="article-body">
            <p>This is the main article content.</p>
            <p>This is another paragraph with important information.</p>
        </div>
        <footer>Site Footer</footer>
    </body>
    </html>`

	result := fetcher.extractTextContent(htmlContent)

	assert.Contains(t, result, "main article content")
	assert.Contains(t, result, "important information")
	assert.NotContains(t, result, "Site Header")
	assert.NotContains(t, result, "Navigation")
	assert.NotContains(t, result, "Site Footer")
}

func TestExtractTextContent_PostBodyPattern(t *testing.T) {
	fetcher := &essayFetcher{}

	htmlContent := `
    <div class="post-body">
        <p>Article content in post body.</p>
        <p>More article text here.</p>
    </div>
    <div class="sidebar">Sidebar content</div>`

	result := fetcher.extractTextContent(htmlContent)

	assert.Contains(t, result, "Article content in post body")
	assert.Contains(t, result, "More article text here")
	assert.NotContains(t, result, "Sidebar content")
}

func TestExtractTextContent_EntryContentPattern(t *testing.T) {
	fetcher := &essayFetcher{}

	htmlContent := `
    <div class="entry-content">
        <p>Blog post content here.</p>
    </div>`

	result := fetcher.extractTextContent(htmlContent)

	assert.Contains(t, result, "Blog post content here")
}

func TestExtractTextContent_ArticleWithHeading(t *testing.T) {
	fetcher := &essayFetcher{}

	htmlContent := `
    <article>
        <h1>Article Title</h1>
        <p>This is the article body content.</p>
        <p>Second paragraph of content.</p>
    </article>`

	result := fetcher.extractTextContent(htmlContent)

	assert.Contains(t, result, "article body content")
	assert.Contains(t, result, "Second paragraph")
	// Title should be excluded by the regex pattern
}

func TestExtractTextContent_FallbackToParagraphs(t *testing.T) {
	fetcher := &essayFetcher{}

	htmlContent := `
    <div class="unknown-structure">
        <h1>Title</h1>
        <p>First paragraph of content.</p>
        <p>Second paragraph with more text.</p>
        <p>Short</p>
        <div>Non-paragraph content</div>
    </div>`

	result := fetcher.extractTextContent(htmlContent)

	assert.Contains(t, result, "First paragraph of content")
	assert.Contains(t, result, "Second paragraph with more text")
	assert.NotContains(t, result, "Short") // Too short, should be filtered
	assert.NotContains(t, result, "Non-paragraph content")
}

func TestExtractParagraphsOnly(t *testing.T) {
	fetcher := &essayFetcher{}

	htmlContent := `
    <div>
        <h1>Article Title</h1>
        <p>This is a good paragraph with enough content.</p>
        <p>Another substantial paragraph here.</p>
        <p>Short</p>
        <p></p>
        <div>Not a paragraph</div>
    </div>`

	result := fetcher.extractParagraphsOnly(htmlContent)

	assert.Contains(t, result, "good paragraph with enough content")
	assert.Contains(t, result, "substantial paragraph here")
	assert.NotContains(t, result, "Short")
	assert.NotContains(t, result, "Article Title")
	assert.NotContains(t, result, "Not a paragraph")
}

func TestCleanTextContent_RemoveScriptAndStyle(t *testing.T) {
	fetcher := &essayFetcher{}

	content := `
    <p>Article content</p>
    <script>alert('test');</script>
    <style>body { color: red; }</style>
    <p>More content</p>`

	result := fetcher.cleanTextContent(content)

	assert.Contains(t, result, "Article content")
	assert.Contains(t, result, "More content")
	assert.NotContains(t, result, "alert")
	assert.NotContains(t, result, "body { color: red; }")
}

func TestCleanTextContent_RemoveHTMLTags(t *testing.T) {
	fetcher := &essayFetcher{}

	content := `<p>Text with <strong>bold</strong> and <em>italic</em> words.</p>`

	result := fetcher.cleanTextContent(content)

	assert.Equal(t, "Text with bold and italic words.", result)
}

func TestCleanTextContent_HTMLEntities(t *testing.T) {
	fetcher := &essayFetcher{}

	content := `<p>Text with &amp; entities &quot;like this&quot; &nbsp; test.</p>`

	result := fetcher.cleanTextContent(content)

	assert.Contains(t, result, "Text with & entities \"like this\" test")
	assert.NotContains(t, result, "&amp;")
	assert.NotContains(t, "&quot;", result)
	assert.NotContains(t, "&nbsp;", result)
}

func TestCleanTextContent_HTMLComments(t *testing.T) {
	fetcher := &essayFetcher{}

	content := `<p>Content before</p><!-- This is a comment --><p>Content after</p>`

	result := fetcher.cleanTextContent(content)

	assert.Contains(t, result, "Content before")
	assert.Contains(t, result, "Content after")
	assert.NotContains(t, result, "This is a comment")
}

func TestCleanTextContent_RemoveAdvertisements(t *testing.T) {
	fetcher := &essayFetcher{}

	content := `<p>Article content Advertisement More content ADVERTISEMENT End</p>`

	result := fetcher.cleanTextContent(content)

	assert.Contains(t, result, "Article content")
	assert.Contains(t, result, "More content")
	assert.Contains(t, result, "End")
	assert.NotContains(t, result, "Advertisement")
	assert.NotContains(t, result, "ADVERTISEMENT")
}

func TestCleanTextContent_NormalizeWhitespace(t *testing.T) {
	fetcher := &essayFetcher{}

	content := `<p>Text   with    multiple     spaces</p>
    <p>And
    newlines</p>`

	result := fetcher.cleanTextContent(content)

	assert.Equal(t, "Text with multiple spaces And newlines", result)
}

func TestExtractTextContent_EmptyContent(t *testing.T) {
	fetcher := &essayFetcher{}

	result := fetcher.extractTextContent("")

	assert.Equal(t, "", result)
}

func TestExtractTextContent_NoMatchingPatterns(t *testing.T) {
	fetcher := &essayFetcher{}

	htmlContent := `
    <div class="weird-structure">
        <span>Some text</span>
        <div>More text in div</div>
    </div>`

	result := fetcher.extractTextContent(htmlContent)

	// Should fallback to extracting paragraphs, but there are none
	// so result should be empty or very minimal
	assert.Equal(t, "", result)
}

func TestExtractTextContent_ComplexHTML(t *testing.T) {
	fetcher := &essayFetcher{}

	htmlContent := `
    <html>
    <head>
        <title>Test Article</title>
        <script>console.log('test');</script>
    </head>
    <body>
        <header>Header content</header>
        <div class="content-body">
            <h1>Article Title</h1>
            <p>First paragraph with <a href="#">links</a> and <strong>formatting</strong>.</p>
            <script>alert('inline script');</script>
            <p>Second paragraph with more substantial content for testing.</p>
            <div class="ad">Advertisement</div>
            <p>Final paragraph after advertisement.</p>
        </div>
        <footer>Footer content</footer>
    </body>
    </html>`

	result := fetcher.extractTextContent(htmlContent)

	assert.Contains(t, result, "First paragraph with links and formatting")
	assert.Contains(t, result, "Second paragraph with more substantial content")
	assert.Contains(t, result, "Final paragraph after advertisement")
	assert.NotContains(t, result, "Header content")
	assert.NotContains(t, result, "Footer content")
	assert.NotContains(t, result, "console.log")
	assert.NotContains(t, result, "alert")
	assert.NotContains(t, result, "Article Title")
}
