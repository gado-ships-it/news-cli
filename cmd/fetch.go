package cmd

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/gado-ships-it/news-cli/internal/ascii"
	"github.com/gado-ships-it/news-cli/internal/config"
	"github.com/gado-ships-it/news-cli/internal/fetcher"
	"github.com/gado-ships-it/news-cli/internal/output"
	"github.com/gado-ships-it/news-cli/internal/seed"
	"github.com/gado-ships-it/news-cli/internal/source"
	"github.com/spf13/cobra"
)

var (
	fetchMax         int
	fetchConcurrency int
	fetchTimeout     time.Duration
)

var fetchCmd = &cobra.Command{
	Use:   "fetch [name...]",
	Short: "Fetch headlines from one or more configured sources.",
	Long: `Fetch headlines from the named sources, or every configured source if
no names are given. Sources can be referenced by their short name
(e.g. "nzz", "nyt", "bbc") — see "news list" for the full set.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := config.Load()
		if err != nil {
			return err
		}
		all := config.Merged(seed.Sources(), store)

		var picked []source.Source
		if len(args) == 0 {
			picked = all
		} else {
			byName := map[string]source.Source{}
			for _, s := range all {
				byName[s.Name] = s
			}
			for _, name := range args {
				s, ok := byName[name]
				if !ok {
					return fmt.Errorf("unknown source %q — try `news list`", name)
				}
				picked = append(picked, s)
			}
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), fetchTimeout)
		defer cancel()
		fps := fetcher.FetchAll(ctx, picked, fetchConcurrency)

		format := output.FromFlags(flagMD, flagJSON)
		asciiFn := buildAsciiFnForFetch(ctx, fps, format)

		return output.Frontpages(os.Stdout, fps, format, fetchMax, asciiFn)
	},
}

// buildAsciiFnForFetch pre-renders ASCII art for every image URL we plan
// to display, then returns a lookup fn. Done eagerly + concurrently so
// rendering doesn't serialize behind the writer.
func buildAsciiFnForFetch(ctx context.Context, fps []source.Frontpage, format output.Format) output.AsciiFn {
	if !flagASCII || format == output.FormatJSON {
		return nil
	}
	type job struct{ url string }
	var jobs []job
	seen := map[string]bool{}
	for _, fp := range fps {
		items := fp.Items
		if fetchMax > 0 && len(items) > fetchMax {
			items = items[:fetchMax]
		}
		for _, it := range items {
			if it.ImageURL == "" || seen[it.ImageURL] {
				continue
			}
			seen[it.ImageURL] = true
			jobs = append(jobs, job{it.ImageURL})
		}
	}
	cache := make(map[string]string, len(jobs))
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, 8)
	for _, j := range jobs {
		wg.Add(1)
		sem <- struct{}{}
		go func(u string) {
			defer wg.Done()
			defer func() { <-sem }()
			art, _ := ascii.Render(ctx, u, flagASCIIWidth)
			if art == "" {
				return
			}
			mu.Lock()
			cache[u] = art
			mu.Unlock()
		}(j.url)
	}
	wg.Wait()
	return func(u string) string { return cache[u] }
}

func init() {
	fetchCmd.Flags().IntVarP(&fetchMax, "max", "n", 10, "max items per source (0 = unlimited)")
	fetchCmd.Flags().IntVar(&fetchConcurrency, "concurrency", 6, "parallel fetches across sources")
	fetchCmd.Flags().DurationVar(&fetchTimeout, "timeout", 30*time.Second, "overall fetch timeout")
}
