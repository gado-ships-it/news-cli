package output

import (
	"fmt"
	"io"
	"os"
)

// ANSI escape building blocks. We use the OSC 8 hyperlink sequence with the
// BEL terminator (\x07) — supported by every modern terminal (Terminal.app,
// iTerm2, Kitty, WezTerm, Ghostty, recent gnome-terminal, Windows Terminal)
// and free of the backslash-escaping pitfalls of the ST form.
const (
	ansiReset     = "\x1b[0m"
	ansiBold      = "\x1b[1m"
	ansiUnderline = "\x1b[4m"
	ansiCyan      = "\x1b[36m"
	ansiBrCyan    = "\x1b[96m"
	ansiBrYellow  = "\x1b[93m"
	ansiGray      = "\x1b[90m" // bright black — renders as a neutral gray on dark and light themes
	osc8Open      = "\x1b]8;;%s\x07"
	osc8Close     = "\x1b]8;;\x07"
)

// styler decides per-call whether to emit escapes. Constructed once per
// render so we don't stat the writer on every line.
type styler struct {
	on bool
}

func newStyler(w io.Writer) *styler {
	return &styler{on: shouldStyle(w)}
}

// shouldStyle returns true when stdout is a TTY and the user hasn't opted
// out via NO_COLOR (https://no-color.org/). FORCE_COLOR=1 bypasses the TTY
// check for testing and for users who pipe through `less -R`.
func shouldStyle(w io.Writer) bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if os.Getenv("FORCE_COLOR") != "" {
		return true
	}
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// headline wraps text in bold + bright cyan.
func (s *styler) headline(text string) string {
	if !s.on {
		return text
	}
	return ansiBold + ansiBrCyan + text + ansiReset
}

// sourceTitle wraps a per-source banner (in fetch) in bold + bright yellow.
func (s *styler) sourceTitle(text string) string {
	if !s.on {
		return text
	}
	return ansiBold + ansiBrYellow + text + ansiReset
}

// link renders label as a clickable terminal hyperlink to url, with gray
// foreground + underline so it reads as a citation rather than primary
// content. When styling is off, label is returned as-is and the URL is
// dropped — callers that need the URL printed separately should do so
// outside this helper.
func (s *styler) link(label, url string) string {
	if !s.on {
		return label
	}
	return fmt.Sprintf(osc8Open+ansiGray+ansiUnderline+"%s"+ansiReset+osc8Close, url, label)
}

// dim wraps text in dim styling for low-emphasis lines (e.g. metadata).
func (s *styler) dim(text string) string {
	if !s.on {
		return text
	}
	return "\x1b[2m" + text + ansiReset
}
