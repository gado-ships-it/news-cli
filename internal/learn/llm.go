package learn

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gado-ships-it/news-cli/internal/source"
)

const (
	maxHTMLChars = 80_000 // keeps the prompt comfortably under any reasonable arg-length cap
)

// Backend wraps a locally-installed coding-agent CLI (Claude Code or Codex)
// and uses it to derive CSS selectors. No API key handling — auth lives in
// whichever CLI the user already has logged in.
type Backend struct {
	Name    string        // "claude" or "codex"
	Bin     string        // resolved absolute path
	Timeout time.Duration
}

// NewBackendFromEnv finds a usable LLM CLI on $PATH.
//
// Preference order: claude, then codex. Override with NEWS_CLI_LLM=claude or
// NEWS_CLI_LLM=codex. Returns an error (not a panic) if neither is present —
// learn's feed/ld+json fallbacks still work without it.
func NewBackendFromEnv() (*Backend, error) {
	candidates := []string{"claude", "codex"}
	if override := strings.TrimSpace(os.Getenv("NEWS_CLI_LLM")); override != "" {
		candidates = []string{override}
	}
	var tried []string
	for _, name := range candidates {
		path, err := exec.LookPath(name)
		if err != nil {
			tried = append(tried, name)
			continue
		}
		return &Backend{Name: name, Bin: path, Timeout: 90 * time.Second}, nil
	}
	return nil, fmt.Errorf("no LLM CLI on $PATH (looked for: %s) — install Claude Code (`claude`) or the OpenAI Codex CLI (`codex`)", strings.Join(tried, ", "))
}

// RunPrompt sends prompt to the backend CLI and returns the raw assistant
// reply as text. Used for any task beyond selector inference (e.g. `news
// tenor` editorial briefs).
//
// Honors the caller's ctx for cancellation/timeout. b.Timeout is applied
// as an additional cap so a runaway subprocess can't outlive the backend's
// configured limit even if the caller forgot to set a ctx deadline.
func (b *Backend) RunPrompt(ctx context.Context, prompt string) (string, error) {
	if b.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, b.Timeout)
		defer cancel()
	}

	cmd, err := b.command(ctx, prompt)
	if err != nil {
		return "", err
	}
	// Run from a neutral working directory so the CLI doesn't accidentally
	// pull in a CLAUDE.md / project context from wherever news-cli was
	// invoked.
	cmd.Dir = os.TempDir()
	cmd.Stdin = bytes.NewReader(nil)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s exec: %w — stderr: %s", b.Name, err, truncate(stderr.String(), 400))
	}
	return stdout.String(), nil
}

// DeriveSelectors asks the backing CLI for a JSON spec describing the
// per-article CSS selectors on a news outlet's homepage.
func (b *Backend) DeriveSelectors(ctx context.Context, homepage string, html []byte) (source.CSSSelectors, error) {
	cleaned := cleanHTML(html)
	if len(cleaned) > maxHTMLChars {
		cleaned = cleaned[:maxHTMLChars]
	}
	prompt := buildPrompt(homepage, cleaned)

	raw, err := b.RunPrompt(ctx, prompt)
	if err != nil {
		return source.CSSSelectors{}, err
	}

	jsonStr := extractJSONObject(raw)
	if jsonStr == "" {
		return source.CSSSelectors{}, fmt.Errorf("%s returned no JSON object — stdout: %s", b.Name, truncate(raw, 400))
	}
	var sel source.CSSSelectors
	if err := json.Unmarshal([]byte(jsonStr), &sel); err != nil {
		return source.CSSSelectors{}, fmt.Errorf("%s returned non-JSON: %w (raw=%q)", b.Name, err, truncate(jsonStr, 400))
	}
	if sel.Item == "" {
		return source.CSSSelectors{}, fmt.Errorf("%s did not produce an item selector", b.Name)
	}
	return sel, nil
}

// command builds the per-backend invocation. Both CLIs accept the prompt as
// a positional argument in their non-interactive modes.
func (b *Backend) command(ctx context.Context, prompt string) (*exec.Cmd, error) {
	switch b.Name {
	case "claude":
		// Claude Code in print mode: one prompt in, one response out.
		// --output-format text returns the raw assistant text (default).
		return exec.CommandContext(ctx, b.Bin,
			"-p", prompt,
			"--output-format", "text",
		), nil
	case "codex":
		// Codex CLI non-interactive exec.
		return exec.CommandContext(ctx, b.Bin, "exec", prompt), nil
	default:
		return nil, fmt.Errorf("unsupported LLM backend %q", b.Name)
	}
}

func buildPrompt(homepage, html string) string {
	return fmt.Sprintf(`You are extracting CSS selectors that identify the article cards on a news outlet's homepage.

Homepage URL: %s

Return STRICT JSON, nothing else, matching this schema exactly:
{"item":"...","headline":"...","link":"...","dek":"...","image":"..."}

Selector rules:
- "item": a selector that matches each repeated article container on the homepage (e.g. "article.story-card", "li[data-testid='story']"). Must repeat 5+ times.
- "headline": selector RELATIVE TO each item, matching the headline element (the text title)
- "link":     selector RELATIVE TO each item, matching the <a> whose href is the article URL
- "dek":      selector RELATIVE TO each item, matching the subtitle/standfirst/abstract. "" if not present.
- "image":    selector RELATIVE TO each item, matching the lede <img>. "" if not present.

Prefer stable, semantic selectors. Avoid Tailwind/MUI hash classes (e.g. "css-1a2b3c"). Prefer aria-* and data-* attributes when stable.

Reply with the JSON object ONLY — no prose, no markdown fences.

HTML follows between <html> tags:
<html>
%s
</html>`, homepage, html)
}

// extractJSONObject finds and returns the first balanced { ... } block in s,
// tolerating leading prose or markdown fences that an LLM may emit.
func extractJSONObject(s string) string {
	start := strings.IndexByte(s, '{')
	if start < 0 {
		return ""
	}
	depth := 0
	for i := start; i < len(s); i++ {
		switch s[i] {
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

// cleanHTML drops noisy nodes so we stay under the token budget without
// losing the structure that drives selector discovery.
func cleanHTML(html []byte) string {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return string(html)
	}
	doc.Find("script, style, svg, noscript, iframe, link, meta").Remove()
	out, err := doc.Html()
	if err != nil {
		return string(html)
	}
	return out
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
