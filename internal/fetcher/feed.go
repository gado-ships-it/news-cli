package fetcher

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/gado-ships-it/news-cli/internal/source"
)

// FetchFeed pulls a Source whose Extractor.Type is ExtractorFeed.
func FetchFeed(ctx context.Context, s source.Source) ([]source.Item, error) {
	if s.Extractor.FeedURL == "" {
		return nil, fmt.Errorf("source %q has no feed_url", s.Name)
	}
	body, _, err := Get(ctx, s.Extractor.FeedURL)
	if err != nil {
		return nil, err
	}
	parser := gofeed.NewParser()
	feed, err := parser.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("parse feed: %w", err)
	}
	now := time.Now().UTC()
	items := make([]source.Item, 0, len(feed.Items))
	for _, fi := range feed.Items {
		it := source.Item{
			Source:    s.Name,
			SourceURL: s.Homepage,
			Headline:  fi.Title,
			Dek:       fi.Description,
			URL:       fi.Link,
			FetchedAt: now,
		}
		if fi.PublishedParsed != nil {
			it.Published = fi.PublishedParsed.UTC()
		} else if fi.UpdatedParsed != nil {
			it.Published = fi.UpdatedParsed.UTC()
		}
		if fi.Image != nil {
			it.ImageURL = fi.Image.URL
		}
		// Some feeds advertise enclosures (e.g. lede image) instead of <image>.
		if it.ImageURL == "" {
			for _, e := range fi.Enclosures {
				if e != nil && e.URL != "" {
					it.ImageURL = e.URL
					break
				}
			}
		}
		items = append(items, it)
	}
	return items, nil
}
