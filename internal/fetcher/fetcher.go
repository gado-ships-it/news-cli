package fetcher

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gado-ships-it/news-cli/internal/source"
)

// Fetch dispatches to the right extractor based on Source.Extractor.Type.
func Fetch(ctx context.Context, s source.Source) source.Frontpage {
	fp := source.Frontpage{Source: s, FetchedAt: time.Now().UTC()}
	var (
		items []source.Item
		err   error
	)
	switch s.Extractor.Type {
	case source.ExtractorFeed:
		items, err = FetchFeed(ctx, s)
	case source.ExtractorLDJSON:
		items, err = FetchLDJSON(ctx, s)
	case source.ExtractorCSS:
		items, err = FetchCSS(ctx, s)
	default:
		err = fmt.Errorf("source %q has no extractor configured — run `news learn %s`", s.Name, s.Homepage)
	}
	if err != nil {
		fp.Error = err.Error()
		return fp
	}
	fp.Items = items
	return fp
}

// FetchAll fans out to multiple sources in parallel with a small worker pool.
func FetchAll(ctx context.Context, sources []source.Source, concurrency int) []source.Frontpage {
	if concurrency < 1 {
		concurrency = 4
	}
	out := make([]source.Frontpage, len(sources))
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	for i, s := range sources {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int, s source.Source) {
			defer wg.Done()
			defer func() { <-sem }()
			out[i] = Fetch(ctx, s)
		}(i, s)
	}
	wg.Wait()
	return out
}
