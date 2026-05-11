package learn

import (
	"bytes"
	"context"
	"errors"
	"net/url"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/mmcdole/gofeed"
	"github.com/gado-ships-it/news-cli/internal/fetcher"
)

// commonFeedPaths is the conventional places publishers stash a feed.
// Ordered roughly by frequency in the wild. These are tried in addition to
// links we discover in the homepage HTML.
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

// preferredFeedHints biases scoring toward "main / all-news" feeds. The more
// of these substrings a candidate URL contains, the higher it ranks. This
// matters because outlets often expose a dozen section feeds (sport,
// finance…) — we want the homepage / top-stories feed.
var preferredFeedHints = []string{
	"startseite", "frontpage", "homepage", // outlet-specific "homepage feed" names
	"recent", "latest", "all", "top", "headlines", "main", "index",
	"news",
}

// indexAnchorHints are URL fragments that suggest a page is itself an
// HTML index listing many feeds — we follow these one level deep.
var indexAnchorHints = []string{
	"/rss", "/rss-feeds", "/feeds", "/feed-list", "/rss/index", "/rss-list",
}

var ErrNoFeed = errors.New("no feed found via autodiscovery, anchor scan, or common paths")

// ProbeFeeds returns the URL of the highest-ranked feed it can verify on the
// outlet's homepage. The strategy is:
//
//  1. Pull the homepage HTML once.
//  2. Collect candidate URLs from:
//     a. <link rel="alternate" type="application/{rss,atom,xml}+xml">
//     b. <a href> matching feed-shaped patterns (*.rss, *.xml, /rss, /feed…)
//     c. The configured commonFeedPaths
//     d. One level of follow-the-index: if the homepage links to an HTML
//        "RSS feeds" overview page, scan that page for further candidates.
//  3. De-dupe, rank by preferredFeedHints, and return the first that parses
//     as a valid RSS/Atom/JSON feed.
func ProbeFeeds(ctx context.Context, homepage string) (string, error) {
	base, err := url.Parse(homepage)
	if err != nil {
		return "", err
	}

	homeBody, _, hErr := fetcher.Get(ctx, homepage)
	var candidates []string

	if hErr == nil {
		candidates = append(candidates, alternateFeedLinks(homeBody, base)...)
		candidates = append(candidates, anchorFeedLinks(homeBody, base)...)

		// Follow obvious "feeds index" anchors one level deep.
		for _, idx := range indexPageAnchors(homeBody, base) {
			body, ct, err := fetcher.Get(ctx, idx)
			if err != nil {
				continue
			}
			// Only walk pages that actually look like HTML.
			if !strings.Contains(strings.ToLower(ct), "text/html") {
				continue
			}
			candidates = append(candidates, anchorFeedLinks(body, base)...)
		}
	}

	for _, p := range commonFeedPaths {
		u := *base
		u.Path = p
		u.RawQuery = ""
		candidates = append(candidates, u.String())
	}

	candidates = dedupe(candidates)
	sortByPreference(candidates)

	parser := gofeed.NewParser()
	for _, c := range candidates {
		body, ct, err := fetcher.Get(ctx, c)
		if err != nil {
			continue
		}
		// Don't waste a parse attempt on responses that announced themselves
		// as HTML — common when a /rss path serves an index page.
		if strings.Contains(strings.ToLower(ct), "text/html") {
			continue
		}
		if _, err := parser.Parse(bytes.NewReader(body)); err == nil {
			return c, nil
		}
	}
	return "", ErrNoFeed
}

// alternateFeedLinks parses <link rel="alternate" type="application/(rss|atom|xml)+xml">.
func alternateFeedLinks(body []byte, base *url.URL) []string {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil
	}
	var out []string
	doc.Find(`link[rel="alternate"]`).Each(func(_ int, s *goquery.Selection) {
		t, _ := s.Attr("type")
		t = strings.ToLower(t)
		if !strings.Contains(t, "rss") && !strings.Contains(t, "atom") && !strings.Contains(t, "xml") && !strings.Contains(t, "feed+json") {
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

// anchorFeedLinks parses <a href> that look feed-shaped. This catches the
// (very common) case where a publisher links to its RSS feeds from the
// footer with plain anchors instead of autodiscovery tags.
func anchorFeedLinks(body []byte, base *url.URL) []string {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil
	}
	var out []string
	doc.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		if href == "" {
			return
		}
		if !looksFeedShaped(href) {
			return
		}
		u, err := url.Parse(href)
		if err != nil {
			return
		}
		// Skip non-http schemes (mailto:, javascript:, …).
		if u.Scheme != "" && u.Scheme != "http" && u.Scheme != "https" {
			return
		}
		out = append(out, base.ResolveReference(u).String())
	})
	return out
}

// indexPageAnchors finds <a href> pointing at likely HTML feed-index pages
// (e.g. /rss, /feeds). One level of follow-the-link.
func indexPageAnchors(body []byte, base *url.URL) []string {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil
	}
	seen := map[string]bool{}
	var out []string
	doc.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		if href == "" {
			return
		}
		u, err := url.Parse(href)
		if err != nil {
			return
		}
		path := strings.ToLower(strings.TrimRight(u.Path, "/"))
		matched := false
		for _, h := range indexAnchorHints {
			if path == h {
				matched = true
				break
			}
		}
		if !matched {
			return
		}
		abs := base.ResolveReference(u).String()
		if seen[abs] {
			return
		}
		seen[abs] = true
		out = append(out, abs)
	})
	return out
}

// looksFeedShaped is a cheap pre-filter — true if the href ends in a feed
// extension or contains a feed-ish path segment.
func looksFeedShaped(href string) bool {
	h := strings.ToLower(href)
	if strings.HasSuffix(h, ".rss") || strings.HasSuffix(h, ".atom") {
		return true
	}
	if strings.HasSuffix(h, ".xml") && (strings.Contains(h, "feed") || strings.Contains(h, "rss") || strings.Contains(h, "atom")) {
		return true
	}
	for _, frag := range []string{"/rss/", "/feed/", "/atom/", "/feeds/", "rss.xml", "feed.xml", "atom.xml"} {
		if strings.Contains(h, frag) {
			return true
		}
	}
	return false
}

// dedupe preserves order while removing duplicates.
func dedupe(in []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(in))
	for _, s := range in {
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}

// sortByPreference is a stable-ish sort that puts URLs containing more
// "main feed" hints first. Ties broken by shorter URL (proxy for less
// section-specific paths).
func sortByPreference(in []string) {
	sort.SliceStable(in, func(i, j int) bool {
		si, sj := score(in[i]), score(in[j])
		if si != sj {
			return si > sj
		}
		return len(in[i]) < len(in[j])
	})
}

func score(u string) int {
	low := strings.ToLower(u)
	n := 0
	for _, h := range preferredFeedHints {
		if strings.Contains(low, h) {
			n++
		}
	}
	return n
}
