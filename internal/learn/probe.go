package learn

import (
	"bytes"
	"context"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/mmcdole/gofeed"
	"github.com/gado-ships-it/news-cli/internal/fetcher"
)

// commonFeedPaths is the conventional places publishers stash a feed.
// Ordered roughly by frequency in the wild.
var commonFeedPaths = []string{
	"/feed",
	"/rss",
	"/rss.xml",
	"/feed.xml",
	"/atom.xml",
	"/index.xml",
	"/feeds/all.atom.xml",
	"/feeds/posts/default",
	"/?feed=rss2",
}

// ProbeFeeds returns the first URL that parses as a valid RSS/Atom feed.
func ProbeFeeds(ctx context.Context, homepage string) (string, error) {
	base, err := url.Parse(homepage)
	if err != nil {
		return "", err
	}
	candidates := candidatesFromHTML(ctx, homepage)
	for _, p := range commonFeedPaths {
		u := *base
		u.Path = p
		u.RawQuery = ""
		candidates = append(candidates, u.String())
	}
	parser := gofeed.NewParser()
	tried := map[string]bool{}
	for _, c := range candidates {
		if tried[c] {
			continue
		}
		tried[c] = true
		body, _, err := fetcher.Get(ctx, c)
		if err != nil {
			continue
		}
		if _, err := parser.Parse(bytes.NewReader(body)); err == nil {
			return c, nil
		}
	}
	return "", errNoFeed
}

var errNoFeed = &probeError{"no feed found via autodiscovery or common paths"}

type probeError struct{ msg string }

func (e *probeError) Error() string { return e.msg }

// candidatesFromHTML pulls <link rel="alternate" type="application/(rss|atom)+xml"> hrefs.
func candidatesFromHTML(ctx context.Context, homepage string) []string {
	body, _, err := fetcher.Get(ctx, homepage)
	if err != nil {
		return nil
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil
	}
	base, _ := url.Parse(homepage)
	var out []string
	doc.Find(`link[rel="alternate"]`).Each(func(_ int, s *goquery.Selection) {
		t, _ := s.Attr("type")
		t = strings.ToLower(t)
		if !strings.Contains(t, "rss") && !strings.Contains(t, "atom") && !strings.Contains(t, "xml") {
			return
		}
		href, ok := s.Attr("href")
		if !ok || href == "" {
			return
		}
		if u, err := url.Parse(href); err == nil && base != nil {
			out = append(out, base.ResolveReference(u).String())
		}
	})
	return out
}
