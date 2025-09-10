package models

import "time"

type Essay struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
	URL     string `json:"url"`
}

type WordCount struct {
	Word  string `json:"word"`
	Count int    `json:"count"`
}

type Output struct {
	TopWords    []WordCount `json:"top_words"`
	TotalEssays int         `json:"total_essays"`
	Timestamp   time.Time   `json:"timestamp"`
}
