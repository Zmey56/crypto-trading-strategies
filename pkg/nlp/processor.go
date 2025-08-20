package nlp

import (
	"strings"
	"unicode"
)

// ProcessedText represents processed text with sentiment analysis
type ProcessedText struct {
	Text          string  `json:"text"`
	Score         float64 `json:"score"`         // Sentiment score from -1.0 to 1.0
	Confidence    float64 `json:"confidence"`    // Confidence level from 0.0 to 1.0
	MentionCount  int     `json:"mention_count"` // Number of mentions
	PositiveWords int     `json:"positive_words"`
	NegativeWords int     `json:"negative_words"`
	NeutralWords  int     `json:"neutral_words"`
}

// Processor handles natural language processing
type Processor struct {
	positiveWords map[string]float64
	negativeWords map[string]float64
	stopWords     map[string]bool
}

// NewProcessor creates a new NLP processor
func NewProcessor() *Processor {
	return &Processor{
		positiveWords: make(map[string]float64),
		negativeWords: make(map[string]float64),
		stopWords:     make(map[string]bool),
	}
}

// ProcessText analyzes sentiment of the given text
func (p *Processor) ProcessText(text string) *ProcessedText {
	// Clean and tokenize text
	tokens := p.tokenize(text)

	// Remove stop words
	filteredTokens := p.removeStopWords(tokens)

	// Count sentiment words
	positiveCount := 0
	negativeCount := 0
	neutralCount := 0

	for _, token := range filteredTokens {
		if p.isPositiveWord(token) {
			positiveCount++
		} else if p.isNegativeWord(token) {
			negativeCount++
		} else {
			neutralCount++
		}
	}

	// Calculate sentiment score
	score := p.calculateSentimentScore(positiveCount, negativeCount, len(filteredTokens))

	// Calculate confidence
	confidence := p.calculateConfidence(positiveCount, negativeCount, neutralCount)

	return &ProcessedText{
		Text:          text,
		Score:         score,
		Confidence:    confidence,
		MentionCount:  len(filteredTokens),
		PositiveWords: positiveCount,
		NegativeWords: negativeCount,
		NeutralWords:  neutralCount,
	}
}

// tokenize splits text into tokens
func (p *Processor) tokenize(text string) []string {
	// Convert to lowercase
	text = strings.ToLower(text)

	// Split by whitespace and punctuation
	fields := strings.FieldsFunc(text, func(r rune) bool {
		return unicode.IsSpace(r) || unicode.IsPunct(r)
	})

	// Clean tokens
	var tokens []string
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if len(field) > 0 {
			tokens = append(tokens, field)
		}
	}

	return tokens
}

// removeStopWords filters out common stop words
func (p *Processor) removeStopWords(tokens []string) []string {
	var filtered []string
	for _, token := range tokens {
		if !p.stopWords[token] {
			filtered = append(filtered, token)
		}
	}
	return filtered
}

// isPositiveWord checks if a word is positive
func (p *Processor) isPositiveWord(word string) bool {
	_, exists := p.positiveWords[word]
	return exists
}

// isNegativeWord checks if a word is negative
func (p *Processor) isNegativeWord(word string) bool {
	_, exists := p.negativeWords[word]
	return exists
}

// calculateSentimentScore calculates sentiment score from -1.0 to 1.0
func (p *Processor) calculateSentimentScore(positive, negative, total int) float64 {
	if total == 0 {
		return 0.0
	}

	positiveRatio := float64(positive) / float64(total)
	negativeRatio := float64(negative) / float64(total)

	return positiveRatio - negativeRatio
}

// calculateConfidence calculates confidence level from 0.0 to 1.0
func (p *Processor) calculateConfidence(positive, negative, neutral int) float64 {
	total := positive + negative + neutral
	if total == 0 {
		return 0.0
	}

	// Higher confidence when there are more sentiment words
	sentimentWords := positive + negative
	return float64(sentimentWords) / float64(total)
}

// AddPositiveWords adds positive words to the dictionary
func (p *Processor) AddPositiveWords(words []string) {
	for _, word := range words {
		p.positiveWords[strings.ToLower(word)] = 1.0
	}
}

// AddNegativeWords adds negative words to the dictionary
func (p *Processor) AddNegativeWords(words []string) {
	for _, word := range words {
		p.negativeWords[strings.ToLower(word)] = 1.0
	}
}

// AddStopWords adds stop words to filter out
func (p *Processor) AddStopWords(words []string) {
	for _, word := range words {
		p.stopWords[strings.ToLower(word)] = true
	}
}

// InitializeDefaultDictionaries initializes with common sentiment words
func (p *Processor) InitializeDefaultDictionaries() {
	// Common positive words
	positiveWords := []string{
		"bull", "bullish", "moon", "pump", "surge", "rally", "gain", "profit",
		"positive", "good", "great", "excellent", "amazing", "wonderful",
		"buy", "long", "hold", "diamond", "hands", "hodl", "lambo",
		"green", "up", "rise", "increase", "growth", "success",
	}

	// Common negative words
	negativeWords := []string{
		"bear", "bearish", "dump", "crash", "drop", "fall", "loss", "red",
		"negative", "bad", "terrible", "awful", "horrible", "worst",
		"sell", "short", "panic", "fear", "scam", "rug", "pull",
		"down", "decline", "decrease", "failure", "dead",
	}

	// Common stop words
	stopWords := []string{
		"the", "a", "an", "and", "or", "but", "in", "on", "at", "to", "for",
		"of", "with", "by", "is", "are", "was", "were", "be", "been", "being",
		"have", "has", "had", "do", "does", "did", "will", "would", "could",
		"should", "may", "might", "can", "this", "that", "these", "those",
		"i", "you", "he", "she", "it", "we", "they", "me", "him", "her",
		"us", "them", "my", "your", "his", "her", "its", "our", "their",
	}

	p.AddPositiveWords(positiveWords)
	p.AddNegativeWords(negativeWords)
	p.AddStopWords(stopWords)
}
