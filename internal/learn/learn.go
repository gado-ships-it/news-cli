package learn

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gado-ships-it/news-cli/internal/fetcher"
	"github.com/gado-ships-it/news-cli/internal/source"
)

// Result is what `news learn` produces: a Source ready to save (with the
// extractor type the learner chose) plus a small validation report.
type Result struct {
	Source     source.Source `json:"source"`
	Method     string        `json:"method"`      // "feed", "ldjson", or "css"
	ItemsFound int           `json:"items_found"` // items returned by the chosen extractor
	Sample     []source.Item `json:"sample"`      // up to 5 items for sanity-checking
}

// Learn runs the discovery pipeline against homepage:
//  1. Try RSS/Atom autodiscovery + common paths
//  2. Try LD-JSON NewsArticle on the homepage
//  3. Ask the LLM for CSS selectors and validate they match
//
// llm may be nil — in that case step 3 is skipped and an error is returned.
func Learn(ctx context.Context, homepage string, llm *Backend) (*Result, error) {
	homepage = strings.TrimSpace(homepage)
	if homepage == "" {
		return nil, fmt.Errorf("homepage required")
	}
	if _, err := url.ParseRequestURI(homepage); err != nil {
		return nil, fmt.Errorf("invalid url: %w", err)
	}

	now := time.Now().UTC()
	src := source.Source{
		Name:      nameFromHomepage(homepage),
		Title:     titleFromHomepage(homepage),
		Homepage:  homepage,
		LearnedAt: &now,
	}

	// 1. Feed
	if feedURL, err := ProbeFeeds(ctx, homepage); err == nil {
		src.Extractor = source.ExtractorSpec{Type: source.ExtractorFeed, FeedURL: feedURL}
		items, ferr := fetcher.FetchFeed(ctx, src)
		if ferr == nil && len(items) > 0 {
			return &Result{Source: src, Method: "feed", ItemsFound: len(items), Sample: sample(items)}, nil
		}
	}

	// 2. LD-JSON
	src.Extractor = source.ExtractorSpec{Type: source.ExtractorLDJSON}
	if items, err := fetcher.FetchLDJSON(ctx, src); err == nil && len(items) > 0 {
		return &Result{Source: src, Method: "ldjson", ItemsFound: len(items), Sample: sample(items)}, nil
	}

	// 3. CSS via LLM
	if llm == nil {
		return nil, fmt.Errorf("no feed or ld+json found; LLM required for CSS fallback (set ANTHROPIC_API_KEY)")
	}
	body, _, err := fetcher.Get(ctx, homepage)
	if err != nil {
		return nil, fmt.Errorf("fetch homepage for selector inference: %w", err)
	}
	sel, err := llm.DeriveSelectors(ctx, homepage, body)
	if err != nil {
		return nil, fmt.Errorf("llm selector inference: %w", err)
	}
	src.Extractor = source.ExtractorSpec{Type: source.ExtractorCSS, Selectors: sel}
	items, err := fetcher.FetchCSS(ctx, src)
	if err != nil {
		return nil, fmt.Errorf("validating inferred selectors: %w (selectors=%+v)", err, sel)
	}
	return &Result{Source: src, Method: "css", ItemsFound: len(items), Sample: sample(items)}, nil
}

func sample(items []source.Item) []source.Item {
	if len(items) > 5 {
		return items[:5]
	}
	return items
}

// nameFromHomepage derives a short, CLI-friendly id from the host.
//
//	https://www.nzz.ch/        -> "nzz"
//	https://www.bbc.com/news   -> "bbc"
func nameFromHomepage(homepage string) string {
	u, err := url.Parse(homepage)
	if err != nil {
		return "source"
	}
	host := strings.TrimPrefix(u.Host, "www.")
	if i := strings.IndexByte(host, '.'); i > 0 {
		return host[:i]
	}
	return host
}

func titleFromHomepage(homepage string) string {
	u, err := url.Parse(homepage)
	if err != nil {
		return homepage
	}
	host := strings.TrimPrefix(u.Host, "www.")
	return host
}
