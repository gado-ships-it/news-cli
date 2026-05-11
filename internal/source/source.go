package source

import "time"

// ExtractorType describes how a Source's frontpage is read.
type ExtractorType string

const (
	ExtractorFeed   ExtractorType = "feed"   // RSS / Atom / JSON Feed
	ExtractorLDJSON ExtractorType = "ldjson" // schema.org NewsArticle in <script type="application/ld+json">
	ExtractorCSS    ExtractorType = "css"    // CSS selectors against rendered HTML
)

// Source is a configured news outlet.
type Source struct {
	Name      string         `json:"name"`        // short id, e.g. "nzz"
	Title     string         `json:"title"`       // display name, e.g. "Neue Zürcher Zeitung"
	Homepage  string         `json:"homepage"`    // canonical homepage URL, always shown for attribution
	Locale    string         `json:"locale"`      // BCP 47, e.g. "de-CH"
	Region    string         `json:"region"`      // free-form, e.g. "Europe"
	Extractor ExtractorSpec  `json:"extractor"`
	LearnedAt *time.Time     `json:"learned_at,omitempty"`
	Notes     string         `json:"notes,omitempty"`
}

// ExtractorSpec is the recipe for turning Source.Homepage into a list of Items.
type ExtractorSpec struct {
	Type ExtractorType `json:"type"`

	// FeedURL is set when Type == ExtractorFeed.
	FeedURL string `json:"feed_url,omitempty"`

	// CSS selectors used when Type == ExtractorCSS.
	// Item is the per-article container; the other selectors are scoped to it.
	Selectors CSSSelectors `json:"selectors,omitempty"`
}

// CSSSelectors are the per-element selectors evaluated relative to Selectors.Item.
type CSSSelectors struct {
	Item     string `json:"item,omitempty"`     // each article container, e.g. "article.story-card"
	Headline string `json:"headline,omitempty"` // headline text inside Item
	Link     string `json:"link,omitempty"`     // <a> whose href is the article URL
	Dek      string `json:"dek,omitempty"`      // subtitle / abstract
	Image    string `json:"image,omitempty"`    // <img> whose src is the lede image
}

// Item is one entry from a Source's frontpage.
type Item struct {
	Source     string    `json:"source"`           // Source.Name — always populated for attribution
	SourceURL  string    `json:"source_url"`       // Source.Homepage
	Headline   string    `json:"headline"`
	Dek        string    `json:"dek,omitempty"`
	URL        string    `json:"url"`              // absolute URL to the article on the source's site
	ImageURL   string    `json:"image_url,omitempty"`
	Published  time.Time `json:"published,omitempty"`
	FetchedAt  time.Time `json:"fetched_at"`
}

// Frontpage is the result of fetching one Source.
type Frontpage struct {
	Source    Source    `json:"source"`
	Items     []Item    `json:"items"`
	FetchedAt time.Time `json:"fetched_at"`
	Error     string    `json:"error,omitempty"`
}

// Brief is the editorial digest produced by `news tenor` — duplicates
// merged, single-event noise stripped, long-lasting stories surfaced. The
// LLM populates Date and Items; GeneratedAt is stamped by the caller.
type Brief struct {
	Date        string      `json:"date"` // YYYY-MM-DD, set by the LLM from the corpus
	Items       []BriefItem `json:"items"`
	GeneratedAt time.Time   `json:"generated_at"`
	Model       string      `json:"model,omitempty"` // backend name (claude / codex)
}

// BriefItem is one merged story in a Brief. Sources is mandatory — every
// item must cite at least one originating outlet for attribution.
type BriefItem struct {
	Headline string         `json:"headline"`
	Summary  string         `json:"summary"`
	Sources  []BriefSource  `json:"sources"`
}

// BriefSource references one outlet's coverage of a BriefItem.
type BriefSource struct {
	Name string `json:"name"` // source short id, e.g. "nzz"
	URL  string `json:"url"`  // article URL on the publisher's site
}
