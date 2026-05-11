package fetcher

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	UserAgent      = "news-cli/0.1 (+https://github.com/gado-ships-it/news-cli; attribution-preserving frontpage aggregator)"
	DefaultTimeout = 20 * time.Second
	MaxBodyBytes   = 8 * 1024 * 1024 // 8 MiB cap to keep memory bounded
)

// Client is the shared HTTP client. Identifies as news-cli with a link so
// publishers can recognize and rate-limit / block as they prefer.
var Client = &http.Client{Timeout: DefaultTimeout}

// Get performs a GET with the standard UA and returns the (capped) body.
func Get(ctx context.Context, url string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Accept", "application/rss+xml, application/atom+xml, application/xml;q=0.9, text/html;q=0.8, */*;q=0.5")
	resp, err := Client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, resp.Header.Get("Content-Type"), fmt.Errorf("http %d for %s", resp.StatusCode, url)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, MaxBodyBytes))
	if err != nil {
		return nil, resp.Header.Get("Content-Type"), err
	}
	return body, resp.Header.Get("Content-Type"), nil
}
