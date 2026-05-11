package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gado-ships-it/news-cli/internal/source"
)

// Format selects between the three rendering modes.
type Format string

const (
	FormatText Format = "text"
	FormatMD   Format = "md"
	FormatJSON Format = "json"
)

// FromFlags picks the right Format from the global --md / --json flags.
// If both are passed --json wins (it's the more structured choice).
func FromFlags(md, jsonOut bool) Format {
	switch {
	case jsonOut:
		return FormatJSON
	case md:
		return FormatMD
	default:
		return FormatText
	}
}

// Frontpages renders a slice of Frontpage in the requested format.
func Frontpages(w io.Writer, fps []source.Frontpage, f Format, maxItems int) error {
	switch f {
	case FormatJSON:
		return writeJSON(w, fps)
	case FormatMD:
		return writeMD(w, fps, maxItems)
	default:
		return writeText(w, fps, maxItems)
	}
}

// Sources renders the configured source list in the requested format.
func Sources(w io.Writer, srcs []source.Source, f Format) error {
	switch f {
	case FormatJSON:
		return writeJSON(w, srcs)
	case FormatMD:
		fmt.Fprintln(w, "# Configured sources")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "| Name | Title | Homepage | Extractor |")
		fmt.Fprintln(w, "|---|---|---|---|")
		for _, s := range srcs {
			t := string(s.Extractor.Type)
			if t == "" {
				t = "_unconfigured — run `news learn`_"
			}
			fmt.Fprintf(w, "| `%s` | %s | [%s](%s) | %s |\n", s.Name, s.Title, s.Homepage, s.Homepage, t)
		}
		return nil
	default:
		for _, s := range srcs {
			t := string(s.Extractor.Type)
			if t == "" {
				t = "unconfigured"
			}
			fmt.Fprintf(w, "%-22s %-40s %s [%s]\n", s.Name, s.Title, s.Homepage, t)
		}
		return nil
	}
}

func writeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func writeMD(w io.Writer, fps []source.Frontpage, maxItems int) error {
	fmt.Fprintf(w, "# News frontpages — fetched %s\n\n", time.Now().UTC().Format(time.RFC3339))
	for _, fp := range fps {
		fmt.Fprintf(w, "## %s\n\n", fp.Source.Title)
		fmt.Fprintf(w, "Source: [%s](%s) · fetched %s\n\n", fp.Source.Homepage, fp.Source.Homepage, fp.FetchedAt.Format(time.RFC3339))
		if fp.Error != "" {
			fmt.Fprintf(w, "> ⚠️ error: %s\n\n", fp.Error)
			continue
		}
		items := fp.Items
		if maxItems > 0 && len(items) > maxItems {
			items = items[:maxItems]
		}
		for _, it := range items {
			fmt.Fprintf(w, "- **[%s](%s)** — _via %s_\n", strings.TrimSpace(it.Headline), it.URL, fp.Source.Title)
			if it.Dek != "" {
				fmt.Fprintf(w, "  > %s\n", oneLine(it.Dek))
			}
			if it.ImageURL != "" {
				fmt.Fprintf(w, "  ![](%s)\n", it.ImageURL)
			}
		}
		fmt.Fprintln(w)
	}
	return nil
}

func writeText(w io.Writer, fps []source.Frontpage, maxItems int) error {
	for _, fp := range fps {
		fmt.Fprintf(w, "=== %s — %s ===\n", fp.Source.Title, fp.Source.Homepage)
		if fp.Error != "" {
			fmt.Fprintf(w, "  ! %s\n\n", fp.Error)
			continue
		}
		items := fp.Items
		if maxItems > 0 && len(items) > maxItems {
			items = items[:maxItems]
		}
		for _, it := range items {
			fmt.Fprintf(w, "  • %s\n", oneLine(it.Headline))
			if it.Dek != "" {
				fmt.Fprintf(w, "    %s\n", oneLine(it.Dek))
			}
			fmt.Fprintf(w, "    %s\n", it.URL)
		}
		fmt.Fprintln(w)
	}
	return nil
}

func oneLine(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return strings.Join(strings.Fields(s), " ")
}

// Brief renders an editorial brief in the requested format.
func Brief(w io.Writer, b *source.Brief, f Format) error {
	switch f {
	case FormatJSON:
		return writeJSON(w, b)
	case FormatMD:
		fmt.Fprintf(w, "# News brief — %s\n\n", b.Date)
		fmt.Fprintf(w, "_Generated %s via %s, dedup'd across configured sources._\n\n",
			b.GeneratedAt.Format(time.RFC3339), b.Model)
		for _, it := range b.Items {
			fmt.Fprintf(w, "## %s\n\n", oneLine(it.Headline))
			fmt.Fprintf(w, "%s\n\n", oneLine(it.Summary))
			if len(it.Sources) > 0 {
				parts := make([]string, 0, len(it.Sources))
				for _, s := range it.Sources {
					parts = append(parts, fmt.Sprintf("[%s](%s)", s.Name, s.URL))
				}
				fmt.Fprintf(w, "Sources: %s\n\n", strings.Join(parts, " · "))
			}
		}
		return nil
	default:
		fmt.Fprintf(w, "News brief — %s  (via %s)\n\n", b.Date, b.Model)
		for _, it := range b.Items {
			fmt.Fprintf(w, "• %s\n", oneLine(it.Headline))
			fmt.Fprintf(w, "  %s\n", oneLine(it.Summary))
			if len(it.Sources) > 0 {
				names := make([]string, 0, len(it.Sources))
				for _, s := range it.Sources {
					names = append(names, s.Name)
				}
				fmt.Fprintf(w, "  sources: %s\n", strings.Join(names, ", "))
				for _, s := range it.Sources {
					fmt.Fprintf(w, "    %s\n", s.URL)
				}
			}
			fmt.Fprintln(w)
		}
		return nil
	}
}
