package fetcher

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gado-ships-it/news-cli/internal/source"
)

// FetchCSS pulls a Source by applying its CSS selector spec to the homepage.
func FetchCSS(ctx context.Context, s source.Source) ([]source.Item, error) {
	sel := s.Extractor.Selectors
	if sel.Item == "" {
		return nil, fmt.Errorf("source %q has no item selector", s.Name)
	}
	body, _, err := Get(ctx, s.Homepage)
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}
	base, _ := url.Parse(s.Homepage)
	now := time.Now().UTC()
	var items []source.Item
	seen := map[string]bool{}
	doc.Find(sel.Item).Each(func(_ int, n *goquery.Selection) {
		it := extractItem(n, sel, s, base, now)
		if it.Headline == "" || it.URL == "" {
			return
		}
		key := it.URL
		if seen[key] {
			return
		}
		seen[key] = true
		items = append(items, it)
	})
	if len(items) == 0 {
		return nil, fmt.Errorf("css selectors matched zero items on %s", s.Homepage)
	}
	return items, nil
}

func extractItem(n *goquery.Selection, sel source.CSSSelectors, s source.Source, base *url.URL, now time.Time) source.Item {
	it := source.Item{
		Source:    s.Name,
		SourceURL: s.Homepage,
		FetchedAt: now,
	}
	if sel.Headline != "" {
		it.Headline = strings.TrimSpace(n.Find(sel.Headline).First().Text())
	}
	if it.Headline == "" {
		// fall back to the item's own text — small CSS-only sites often have <a>headline</a>
		it.Headline = strings.TrimSpace(n.Text())
	}
	if sel.Link != "" {
		if href, ok := n.Find(sel.Link).First().Attr("href"); ok {
			it.URL = absURL(base, href)
		}
	}
	if it.URL == "" {
		if href, ok := n.Find("a").First().Attr("href"); ok {
			it.URL = absURL(base, href)
		}
	}
	if sel.Dek != "" {
		it.Dek = strings.TrimSpace(n.Find(sel.Dek).First().Text())
	}
	if sel.Image != "" {
		img := n.Find(sel.Image).First()
		if src, ok := img.Attr("src"); ok && src != "" {
			it.ImageURL = absURL(base, src)
		} else if src, ok := img.Attr("data-src"); ok && src != "" {
			it.ImageURL = absURL(base, src)
		} else if srcset, ok := img.Attr("srcset"); ok && srcset != "" {
			it.ImageURL = absURL(base, firstSrcsetURL(srcset))
		}
	}
	return it
}

func firstSrcsetURL(srcset string) string {
	parts := strings.Split(srcset, ",")
	if len(parts) == 0 {
		return ""
	}
	first := strings.TrimSpace(parts[0])
	fields := strings.Fields(first)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}
