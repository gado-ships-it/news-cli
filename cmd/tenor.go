package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gado-ships-it/news-cli/internal/ascii"
	"github.com/gado-ships-it/news-cli/internal/config"
	"github.com/gado-ships-it/news-cli/internal/fetcher"
	"github.com/gado-ships-it/news-cli/internal/learn"
	"github.com/gado-ships-it/news-cli/internal/output"
	"github.com/gado-ships-it/news-cli/internal/seed"
	"github.com/gado-ships-it/news-cli/internal/source"
	"github.com/spf13/cobra"
)

var (
	tenorMax         int
	tenorConcurrency int
	tenorTimeout     time.Duration
	tenorEntries     int
)

var tenorCmd = &cobra.Command{
	Use:   "tenor [name...]",
	Short: "Brief of today's events: deduped, long-lasting, non-clickbait.",
	Long: `tenor fetches the configured frontpages, then asks your locally-installed
LLM CLI (claude or codex) to produce a short editorial brief.

The brief deduplicates stories covered by multiple outlets, de-emphasizes
single-event human-interest pieces, and prioritizes news with long-term
consequence — geopolitics, policy, science, structural economic shifts,
durable technology trends, climate. Every entry cites its originating
outlet(s) with article URLs.

Requires claude or codex on $PATH (set NEWS_CLI_LLM to force one).`,
	RunE: runTenor,
}

func init() {
	tenorCmd.Flags().IntVarP(&tenorMax, "max", "n", 10, "max items per source sent to the LLM")
	tenorCmd.Flags().IntVar(&tenorEntries, "entries", 8, "target number of merged entries in the brief")
	tenorCmd.Flags().IntVar(&tenorConcurrency, "concurrency", 6, "parallel fetches across sources")
	tenorCmd.Flags().DurationVar(&tenorTimeout, "timeout", 5*time.Minute, "overall fetch+LLM timeout")
}

func runTenor(cmd *cobra.Command, args []string) error {
	store, err := config.Load()
	if err != nil {
		return err
	}
	all := config.Merged(seed.Sources(), store)

	picked, err := pickSources(all, args)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(cmd.Context(), tenorTimeout)
	defer cancel()

	fmt.Fprintf(os.Stderr, "fetching %d sources…\n", len(picked))
	fps := fetcher.FetchAll(ctx, picked, tenorConcurrency)

	corpus := buildTenorCorpus(fps, tenorMax)
	if totalCorpusItems(corpus) == 0 {
		return fmt.Errorf("no items fetched — every source errored. check `news list` and your network")
	}

	backend, err := learn.NewBackendFromEnv()
	if err != nil {
		return fmt.Errorf("tenor requires an LLM CLI: %w", err)
	}
	// Tenor prompts can be large; give the backend a longer ceiling than
	// learn (which uses 90s) so editorial reasoning has time to finish.
	backend.Timeout = tenorTimeout

	fmt.Fprintf(os.Stderr, "asking %s for a brief over %d headlines from %d sources…\n",
		backend.Name, totalCorpusItems(corpus), len(corpus))

	prompt := buildTenorPrompt(corpus, tenorEntries)
	raw, err := backend.RunPrompt(ctx, prompt)
	if err != nil {
		return err
	}

	brief, err := parseBrief(raw)
	if err != nil {
		return fmt.Errorf("LLM returned unparseable brief: %w\n--- raw (first 400 chars) ---\n%s",
			err, truncForErr(raw, 400))
	}
	brief.GeneratedAt = time.Now().UTC()
	brief.Model = backend.Name
	if brief.Date == "" {
		brief.Date = time.Now().UTC().Format("2006-01-02")
	}

	format := output.FromFlags(flagMD, flagJSON)
	asciiFn := buildAsciiFnForTenor(ctx, fps, brief, format)

	return output.Brief(os.Stdout, brief, format, asciiFn)
}

// buildAsciiFnForTenor maps each brief item's first source article URL to
// the ASCII art rendered from that original Item's ImageURL. Two-step
// lookup because brief items don't carry image URLs; the fetched Items do.
func buildAsciiFnForTenor(ctx context.Context, fps []source.Frontpage, brief *source.Brief, format output.Format) output.AsciiFn {
	if !flagASCII || format == output.FormatJSON {
		return nil
	}
	// article URL -> image URL, from the fetched corpus
	imgByArticle := map[string]string{}
	for _, fp := range fps {
		for _, it := range fp.Items {
			if it.URL != "" && it.ImageURL != "" {
				imgByArticle[it.URL] = it.ImageURL
			}
		}
	}
	// Collect the image URLs we actually need (one per brief item).
	want := map[string]string{} // articleURL -> imageURL
	for _, bi := range brief.Items {
		for _, s := range bi.Sources {
			if img, ok := imgByArticle[s.URL]; ok && img != "" {
				want[s.URL] = img
				break
			}
		}
	}
	if len(want) == 0 {
		return nil
	}
	cache := make(map[string]string, len(want))
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, 6)
	for articleURL, imgURL := range want {
		wg.Add(1)
		sem <- struct{}{}
		go func(art, img string) {
			defer wg.Done()
			defer func() { <-sem }()
			rendered, _ := ascii.Render(ctx, img, flagASCIIWidth)
			if rendered == "" {
				return
			}
			mu.Lock()
			cache[art] = rendered
			mu.Unlock()
		}(articleURL, imgURL)
	}
	wg.Wait()
	return func(articleURL string) string { return cache[articleURL] }
}

// pickSources resolves CLI arg names against the merged source list. Empty
// args means "all configured sources".
func pickSources(all []source.Source, args []string) ([]source.Source, error) {
	if len(args) == 0 {
		return all, nil
	}
	byName := map[string]source.Source{}
	for _, s := range all {
		byName[s.Name] = s
	}
	var out []source.Source
	for _, name := range args {
		s, ok := byName[name]
		if !ok {
			return nil, fmt.Errorf("unknown source %q — try `news list`", name)
		}
		out = append(out, s)
	}
	return out, nil
}

// corpusEntry is one source's contribution to the brief prompt. Trimmed
// down to what the LLM actually needs.
type corpusEntry struct {
	Source string         `json:"source"`
	Title  string         `json:"title"`
	URL    string         `json:"url"`
	Items  []corpusItem   `json:"items"`
}

type corpusItem struct {
	Headline  string `json:"headline"`
	Dek       string `json:"dek,omitempty"`
	URL       string `json:"url"`
	Published string `json:"published,omitempty"`
}

func buildTenorCorpus(fps []source.Frontpage, maxPerSource int) []corpusEntry {
	out := make([]corpusEntry, 0, len(fps))
	for _, fp := range fps {
		if fp.Error != "" || len(fp.Items) == 0 {
			continue
		}
		items := fp.Items
		if maxPerSource > 0 && len(items) > maxPerSource {
			items = items[:maxPerSource]
		}
		ce := corpusEntry{
			Source: fp.Source.Name,
			Title:  fp.Source.Title,
			URL:    fp.Source.Homepage,
		}
		for _, it := range items {
			ci := corpusItem{
				Headline: oneLineTen(it.Headline),
				Dek:      truncTen(oneLineTen(it.Dek), 280),
				URL:      it.URL,
			}
			if !it.Published.IsZero() {
				ci.Published = it.Published.Format(time.RFC3339)
			}
			ce.Items = append(ce.Items, ci)
		}
		out = append(out, ce)
	}
	return out
}

func totalCorpusItems(corpus []corpusEntry) int {
	n := 0
	for _, c := range corpus {
		n += len(c.Items)
	}
	return n
}

func buildTenorPrompt(corpus []corpusEntry, entries int) string {
	corpusJSON, _ := json.Marshal(corpus)
	today := time.Now().UTC().Format("2006-01-02")
	return fmt.Sprintf(`You are a constructive news editor. Below is a JSON list of frontpage headlines fetched today from multiple major news outlets.

Produce a short editorial BRIEF of today's events.

Rules:
1. DEDUPE: when multiple outlets cover the same underlying story, merge them into ONE entry and list all the originating outlets in "sources".
2. PRIORITIZE long-term consequence: geopolitics, policy, durable economic shifts, scientific results, structural technology trends, climate, institutional events, public health.
3. DE-EMPHASIZE (usually omit): celebrity gossip, sports scores, single-event human-interest, weather, lifestyle, daily market ticker noise, clickbait framings.
4. Target %d merged entries. Each summary 1–2 neutral sentences. Headlines should be informative, not provocative.
5. Every entry MUST cite at least one source with the article URL taken VERBATIM from the corpus. Never invent URLs.
6. If multiple languages are present, write the brief in English but preserve key proper nouns.

Reply with STRICT JSON ONLY — no prose, no markdown fences:

{
  "date": "%s",
  "items": [
    {
      "headline": "Short factual headline of the merged story",
      "summary": "1–2 sentence neutral summary",
      "sources": [
        {"name": "<source short id>", "url": "<article URL exactly as in corpus>"}
      ]
    }
  ]
}

Today's date is %s. Today's frontpage corpus follows:

%s`, entries, today, today, string(corpusJSON))
}

// parseBrief locates the first balanced JSON object in raw and unmarshals
// it into a Brief. Tolerates leading prose or markdown fences.
func parseBrief(raw string) (*source.Brief, error) {
	jsonStr := extractFirstJSONObject(raw)
	if jsonStr == "" {
		return nil, fmt.Errorf("no JSON object found in LLM reply")
	}
	var b source.Brief
	if err := json.Unmarshal([]byte(jsonStr), &b); err != nil {
		return nil, err
	}
	if len(b.Items) == 0 {
		return nil, fmt.Errorf("brief contains zero items")
	}
	return &b, nil
}

func extractFirstJSONObject(s string) string {
	start := strings.IndexByte(s, '{')
	if start < 0 {
		return ""
	}
	depth := 0
	inString := false
	escape := false
	for i := start; i < len(s); i++ {
		c := s[i]
		if escape {
			escape = false
			continue
		}
		if inString {
			switch c {
			case '\\':
				escape = true
			case '"':
				inString = false
			}
			continue
		}
		switch c {
		case '"':
			inString = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return s[start : i+1]
			}
		}
	}
	return ""
}

func oneLineTen(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return strings.Join(strings.Fields(s), " ")
}

func truncTen(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func truncForErr(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
