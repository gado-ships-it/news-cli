// Package ascii renders an image URL as a block of ASCII characters sized
// for the terminal. Tiny, dep-free implementation — fetches via the shared
// HTTP client, decodes anything stdlib supports (PNG/JPEG/GIF), nearest-
// neighbor downsamples to the target character grid, and maps Rec. 601
// luminance to a 10-step glyph ramp.
package ascii

import (
	"bytes"
	"context"
	"fmt"
	"image"
	// Register stdlib decoders for side-effects so image.Decode handles
	// every format publishers actually serve in <meta property="og:image">.
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"strings"

	"github.com/gado-ships-it/news-cli/internal/fetcher"
)

// ramp goes light → dark. 10 levels balances detail and readability;
// terminals tend to muddy denser ramps unless you use truecolor.
const ramp = " .:-=+*#%@"

// terminalCellAspect is the ratio of character-cell height to width
// (typical monospace fonts are roughly twice as tall as they are wide).
// We divide image rows by this so squares come out square on screen.
const terminalCellAspect = 2.0

// Render downloads imageURL and returns it as a width × derivedHeight grid
// of ramp characters with newlines between rows. width is in characters;
// height is computed from the source aspect ratio. Returns ("", err) on
// any failure — callers typically fall back to skipping the image rather
// than surfacing the error.
func Render(ctx context.Context, imageURL string, width int) (string, error) {
	if imageURL == "" {
		return "", fmt.Errorf("ascii: no image url")
	}
	if width <= 0 {
		width = 60
	}
	body, _, err := fetcher.Get(ctx, imageURL)
	if err != nil {
		return "", fmt.Errorf("ascii: fetch %s: %w", imageURL, err)
	}
	img, _, err := image.Decode(bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("ascii: decode %s: %w", imageURL, err)
	}
	return renderImage(img, width), nil
}

func renderImage(img image.Image, width int) string {
	b := img.Bounds()
	iw := b.Dx()
	ih := b.Dy()
	if iw == 0 || ih == 0 {
		return ""
	}
	height := int(float64(width) * float64(ih) / float64(iw) / terminalCellAspect)
	if height < 1 {
		height = 1
	}

	var sb strings.Builder
	sb.Grow((width + 1) * height)
	for y := 0; y < height; y++ {
		srcY := b.Min.Y + y*ih/height
		for x := 0; x < width; x++ {
			srcX := b.Min.X + x*iw/width
			r, g, bl, a := img.At(srcX, srcY).RGBA()
			// Transparent pixels render as blanks so logo backgrounds
			// don't bleed into a uniform dark block.
			if a < 0x4000 {
				sb.WriteByte(' ')
				continue
			}
			// .RGBA() returns 16-bit components; collapse to 8.
			rr := float64(r >> 8)
			gg := float64(g >> 8)
			bb := float64(bl >> 8)
			lum := 0.299*rr + 0.587*gg + 0.114*bb // 0..255
			idx := int(lum / 256.0 * float64(len(ramp)))
			if idx >= len(ramp) {
				idx = len(ramp) - 1
			}
			sb.WriteByte(ramp[idx])
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}
