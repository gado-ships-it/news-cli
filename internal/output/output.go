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
// AsciiFn returns rendered ASCII art for a given URL, or "" to skip.
// JSON output ignores it; text and md modes embed the art under the
// matching item. The key is whichever URL space the caller wants to look
// up — typically Item.ImageURL for frontpages and the brief item's first
// source article URL for tenor.
type AsciiFn func(url string) string

func Frontpages(w io.Writer, fps []source.Frontpage, f Format, maxItems int, asciiByImageURL AsciiFn) error {
	switch f {
	case FormatJSON:
		return writeJSON(w, fps)
	case FormatMD:
		return writeMD(w, fps, maxItems, asciiByImageURL)
	default:
		return writeText(w, fps, maxItems, asciiByImageURL)
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

func writeMD(w io.Writer, fps []source.Frontpage, maxItems int, asciiFn AsciiFn) error {
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
			if asciiFn != nil && it.ImageURL != "" {
				if art := asciiFn(it.ImageURL); art != "" {
					fmt.Fprintln(w, "  ```")
					fmt.Fprint(w, indent(art, "  "))
					fmt.Fprintln(w, "  ```")
					continue
				}
			}
			if it.ImageURL != "" {
				fmt.Fprintf(w, "  ![](%s)\n", it.ImageURL)
			}
		}
		fmt.Fprintln(w)
	}
	return nil
}

func writeText(w io.Writer, fps []source.Frontpage, maxItems int, asciiFn AsciiFn) error {
	st := newStyler(w)
	for _, fp := range fps {
		// Source banner: title in bold yellow, homepage as a dim hyperlink.
		if st.on {
			fmt.Fprintf(w, "=== %s · %s ===\n", st.sourceTitle(fp.Source.Title), st.link(fp.Source.Homepage, fp.Source.Homepage))
		} else {
			fmt.Fprintf(w, "=== %s — %s ===\n", fp.Source.Title, fp.Source.Homepage)
		}
		if fp.Error != "" {
			fmt.Fprintf(w, "  ! %s\n\n", fp.Error)
			continue
		}
		items := fp.Items
		if maxItems > 0 && len(items) > maxItems {
			items = items[:maxItems]
		}
		for _, it := range items {
			fmt.Fprintf(w, "  • %s\n", st.headline(oneLine(it.Headline)))
			if it.Dek != "" {
				fmt.Fprintf(w, "    %s\n", oneLine(it.Dek))
			}
			// URL line: still printed in both modes, but on a TTY it's
			// gray + underlined + clickable.
			if it.URL != "" {
				fmt.Fprintf(w, "    %s\n", st.link(it.URL, it.URL))
			}
			if asciiFn != nil && it.ImageURL != "" {
				if art := asciiFn(it.ImageURL); art != "" {
					fmt.Fprint(w, indent(art, "    "))
				}
			}
		}
		fmt.Fprintln(w)
	}
	return nil
}

// firstNonEmpty returns the ASCII art for the first BriefSource whose URL
// the renderer can resolve. Used when a tenor entry merges multiple
// outlets and we just want a single representative thumbnail.
func firstNonEmpty(fn AsciiFn, sources []source.BriefSource) string {
	for _, s := range sources {
		if art := fn(s.URL); art != "" {
			return art
		}
	}
	return ""
}

// indent prefixes every non-empty line with prefix. Used to align embedded
// ASCII art under the item it belongs to.
func indent(s, prefix string) string {
	if s == "" {
		return s
	}
	lines := strings.Split(s, "\n")
	for i, ln := range lines {
		if ln == "" {
			continue
		}
		lines[i] = prefix + ln
	}
	return strings.Join(lines, "\n")
}

func oneLine(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return strings.Join(strings.Fields(s), " ")
}

// Brief renders an editorial brief in the requested format. asciiFn is
// keyed on the article URL of a brief item's first source — the tenor
// command builds that map from its fetched frontpages.
func Brief(w io.Writer, b *source.Brief, f Format, asciiFn AsciiFn) error {
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
			if asciiFn != nil && len(it.Sources) > 0 {
				if art := firstNonEmpty(asciiFn, it.Sources); art != "" {
					fmt.Fprintln(w, "```")
					fmt.Fprint(w, art)
					fmt.Fprintln(w, "```")
					fmt.Fprintln(w)
				}
			}
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
		st := newStyler(w)
		fmt.Fprintf(w, "News brief — %s  %s\n\n", b.Date, st.dim("(via "+b.Model+")"))
		for _, it := range b.Items {
			fmt.Fprintf(w, "• %s\n", st.headline(oneLine(it.Headline)))
			fmt.Fprintf(w, "  %s\n", oneLine(it.Summary))
			if asciiFn != nil && len(it.Sources) > 0 {
				if art := firstNonEmpty(asciiFn, it.Sources); art != "" {
					fmt.Fprint(w, indent(art, "  "))
				}
			}
			if len(it.Sources) > 0 {
				parts := make([]string, 0, len(it.Sources))
				for _, s := range it.Sources {
					parts = append(parts, st.link(s.Name, s.URL))
				}
				if st.on {
					// On a TTY: source names are the only thing shown; each
					// is gray + underlined + clickable via OSC 8.
					fmt.Fprintf(w, "  %s %s\n", st.dim("sources:"), strings.Join(parts, st.dim(", ")))
				} else {
					// Off TTY: keep the URLs visible so scripts can grep.
					names := make([]string, 0, len(it.Sources))
					for _, s := range it.Sources {
						names = append(names, s.Name)
					}
					fmt.Fprintf(w, "  sources: %s\n", strings.Join(names, ", "))
					for _, s := range it.Sources {
						fmt.Fprintf(w, "    %s\n", s.URL)
					}
				}
			}
			fmt.Fprintln(w)
		}
		return nil
	}
}
