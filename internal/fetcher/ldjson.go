package fetcher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gado-ships-it/news-cli/internal/source"
)

// FetchLDJSON pulls a Source by walking <script type="application/ld+json">
// blocks on the homepage and harvesting schema.org NewsArticle / ItemList
// entries.
func FetchLDJSON(ctx context.Context, s source.Source) ([]source.Item, error) {
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
	doc.Find(`script[type="application/ld+json"]`).Each(func(_ int, sel *goquery.Selection) {
		raw := strings.TrimSpace(sel.Text())
		if raw == "" {
			return
		}
		harvestLDJSON(raw, s, base, now, &items)
	})
	if len(items) == 0 {
		return nil, fmt.Errorf("no NewsArticle / ItemList found in ld+json")
	}
	return items, nil
}

func harvestLDJSON(raw string, s source.Source, base *url.URL, now time.Time, out *[]source.Item) {
	var node any
	if err := json.Unmarshal([]byte(raw), &node); err != nil {
		return
	}
	walkLD(node, s, base, now, out)
}

func walkLD(node any, s source.Source, base *url.URL, now time.Time, out *[]source.Item) {
	switch v := node.(type) {
	case map[string]any:
		// Handle @graph wrappers.
		if g, ok := v["@graph"]; ok {
			walkLD(g, s, base, now, out)
		}
		t := ldTypes(v["@type"])
		if containsAny(t, "NewsArticle", "ReportageNewsArticle", "AnalysisNewsArticle", "OpinionNewsArticle", "Article") {
			if it, ok := itemFromArticle(v, s, base, now); ok {
				*out = append(*out, it)
			}
		}
		if containsAny(t, "ItemList") {
			if elems, ok := v["itemListElement"].([]any); ok {
				for _, e := range elems {
					if m, ok := e.(map[string]any); ok {
						if inner, ok := m["item"]; ok {
							walkLD(inner, s, base, now, out)
						} else {
							walkLD(m, s, base, now, out)
						}
					}
				}
			}
		}
	case []any:
		for _, x := range v {
			walkLD(x, s, base, now, out)
		}
	}
}

func itemFromArticle(m map[string]any, s source.Source, base *url.URL, now time.Time) (source.Item, bool) {
	headline := str(m["headline"])
	if headline == "" {
		headline = str(m["name"])
	}
	if headline == "" {
		return source.Item{}, false
	}
	it := source.Item{
		Source:    s.Name,
		SourceURL: s.Homepage,
		Headline:  headline,
		Dek:       str(m["description"]),
		URL:       absURL(base, str(m["url"])),
		FetchedAt: now,
	}
	if img := str(m["image"]); img != "" {
		it.ImageURL = absURL(base, img)
	} else if imgs, ok := m["image"].([]any); ok && len(imgs) > 0 {
		it.ImageURL = absURL(base, str(imgs[0]))
	} else if im, ok := m["image"].(map[string]any); ok {
		it.ImageURL = absURL(base, str(im["url"]))
	}
	if d := str(m["datePublished"]); d != "" {
		if t, err := time.Parse(time.RFC3339, d); err == nil {
			it.Published = t.UTC()
		}
	}
	return it, true
}

func ldTypes(v any) []string {
	switch t := v.(type) {
	case string:
		return []string{t}
	case []any:
		out := make([]string, 0, len(t))
		for _, x := range t {
			if s, ok := x.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}

func containsAny(haystack []string, needles ...string) bool {
	for _, h := range haystack {
		for _, n := range needles {
			if h == n {
				return true
			}
		}
	}
	return false
}

func str(v any) string {
	s, _ := v.(string)
	return s
}

func absURL(base *url.URL, raw string) string {
	if raw == "" || base == nil {
		return raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	return base.ResolveReference(u).String()
}
