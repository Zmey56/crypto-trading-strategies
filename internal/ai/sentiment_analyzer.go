package ai

import (
	"context"
	"time"

	"github.com/Zmey56/crypto-arbitrage-trader/pkg/nlp"
)

type SentimentAnalyzer struct {
	nlpProcessor *nlp.Processor
	dataSources  map[string]DataSource
	aggregator   *SentimentAggregator
}

type SentimentData struct {
	Source     string    `json:"source"`
	Symbol     string    `json:"symbol"`
	Sentiment  float64   `json:"sentiment"` // -1.0 to 1.0
	Confidence float64   `json:"confidence"`
	Timestamp  time.Time `json:"timestamp"`
	Volume     int       `json:"mention_volume"`
}

// AnalyzeMarketSentiment processes multiple data sources concurrently
func (sa *SentimentAnalyzer) AnalyzeMarketSentiment(
	ctx context.Context,
	symbol string,
	timeframe time.Duration,
) (*AggregatedSentiment, error) {

	var sentiments []SentimentData

	// Process multiple sources in parallel
	for sourceName, source := range sa.dataSources {
		go func(name string, src DataSource) {
			data, err := src.FetchData(ctx, symbol, timeframe)
			if err != nil {
				return
			}

			processed := sa.nlpProcessor.ProcessText(data)
			sentiment := SentimentData{
				Source:     name,
				Symbol:     symbol,
				Sentiment:  processed.Score,
				Confidence: processed.Confidence,
				Timestamp:  time.Now(),
				Volume:     processed.MentionCount,
			}

			sentiments = append(sentiments, sentiment)
		}(sourceName, source)
	}

	return sa.aggregator.Aggregate(sentiments), nil
}

type DataSource interface {
	FetchData(ctx context.Context, symbol string, timeframe time.Duration) ([]string, error)
}

// TwitterSource analyzes Twitter/X data
type TwitterSource struct {
	apiClient *TwitterAPI
}

// NewsSource processes financial news
type NewsSource struct {
	feeds []NewsFeed
}

// RedditSource analyzes Reddit discussions
type RedditSource struct {
	subreddits []string
}
