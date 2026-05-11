package seed

import "github.com/gado-ships-it/news-cli/internal/source"

// Sources returns the baked-in starter list. Names are stable short ids
// usable on the CLI (e.g. `news fetch nzz`).
//
// The list combines the outlets used by tenor.news with the major US/UK
// outlets from pippinlee/news-cli's pre-existing CLI source list. Feed URLs
// were chosen from each outlet's publicly advertised RSS/Atom endpoints —
// where a publisher does not advertise a feed (e.g. African Intelligence),
// the source is marked with Extractor.Type "" and must be completed via
// `news learn`.
func Sources() []source.Source {
	return []source.Source{
		// --- tenor.news lineup ---
		{
			Name: "semafor", Title: "Semafor", Homepage: "https://www.semafor.com/", Locale: "en-US", Region: "Global",
			Notes: "No public RSS feed advertised; run `news learn https://www.semafor.com/` to derive selectors.",
		},
		{
			Name: "nzz", Title: "Neue Zürcher Zeitung", Homepage: "https://www.nzz.ch/", Locale: "de-CH", Region: "Europe",
			Extractor: source.ExtractorSpec{Type: source.ExtractorFeed, FeedURL: "https://www.nzz.ch/recent.rss"},
		},
		{
			Name: "nyt", Title: "The New York Times", Homepage: "https://www.nytimes.com/", Locale: "en-US", Region: "North America",
			Extractor: source.ExtractorSpec{Type: source.ExtractorFeed, FeedURL: "https://rss.nytimes.com/services/xml/rss/nyt/HomePage.xml"},
		},
		{
			Name: "economist", Title: "The Economist", Homepage: "https://www.economist.com/", Locale: "en-GB", Region: "Global",
			Extractor: source.ExtractorSpec{Type: source.ExtractorFeed, FeedURL: "https://www.economist.com/the-world-this-week/rss.xml"},
		},
		{
			Name: "foreign-policy", Title: "Foreign Policy", Homepage: "https://foreignpolicy.com/", Locale: "en-US", Region: "Global",
			Extractor: source.ExtractorSpec{Type: source.ExtractorFeed, FeedURL: "https://foreignpolicy.com/feed/"},
		},
		{
			Name: "african-arguments", Title: "African Arguments", Homepage: "https://africanarguments.org/", Locale: "en", Region: "Africa",
			Extractor: source.ExtractorSpec{Type: source.ExtractorFeed, FeedURL: "https://africanarguments.org/feed"},
		},
		{
			Name: "lemonde", Title: "Le Monde", Homepage: "https://www.lemonde.fr/", Locale: "fr-FR", Region: "Europe",
			Extractor: source.ExtractorSpec{Type: source.ExtractorFeed, FeedURL: "https://www.lemonde.fr/rss/une.xml"},
		},
		{
			Name: "korea-herald", Title: "The Korea Herald", Homepage: "https://www.koreaherald.com/", Locale: "en-KR", Region: "Asia",
			Extractor: source.ExtractorSpec{Type: source.ExtractorFeed, FeedURL: "https://www.koreaherald.com/rss/020000000000.xml"},
		},
		{
			Name: "cna", Title: "CNA — Channel News Asia", Homepage: "https://www.channelnewsasia.com/", Locale: "en-SG", Region: "Asia",
			Extractor: source.ExtractorSpec{Type: source.ExtractorFeed, FeedURL: "https://www.channelnewsasia.com/api/v1/rss-outbound-feed?_format=xml"},
		},
		{
			Name: "nature", Title: "Nature", Homepage: "https://www.nature.com/", Locale: "en", Region: "Global",
			Extractor: source.ExtractorSpec{Type: source.ExtractorFeed, FeedURL: "https://www.nature.com/nature.rss"},
		},
		{
			Name: "tsuri", Title: "Tsüri.ch", Homepage: "https://tsri.ch/", Locale: "de-CH", Region: "Europe",
			Extractor: source.ExtractorSpec{Type: source.ExtractorFeed, FeedURL: "https://tsri.ch/api/rss-feed"},
		},
		{
			Name: "monde-diplo", Title: "Le Monde diplomatique", Homepage: "https://mondediplo.com/", Locale: "fr", Region: "Global",
			Extractor: source.ExtractorSpec{Type: source.ExtractorFeed, FeedURL: "https://mondediplo.com/spip.php?page=backend"},
		},
		{
			Name: "srf", Title: "SRF — Schweizer Radio und Fernsehen", Homepage: "https://www.srf.ch/news", Locale: "de-CH", Region: "Europe",
			Extractor: source.ExtractorSpec{Type: source.ExtractorFeed, FeedURL: "https://www.srf.ch/news/bnf/rss/1646"},
		},
		{
			Name: "republik", Title: "Republik", Homepage: "https://www.republik.ch/", Locale: "de-CH", Region: "Europe",
			Extractor: source.ExtractorSpec{Type: source.ExtractorFeed, FeedURL: "https://www.republik.ch/feed.xml"},
		},
		{
			Name: "bbc", Title: "BBC News", Homepage: "https://www.bbc.com/news", Locale: "en-GB", Region: "Global",
			Extractor: source.ExtractorSpec{Type: source.ExtractorFeed, FeedURL: "https://feeds.bbci.co.uk/news/rss.xml"},
		},
		{
			Name: "wapo", Title: "The Washington Post", Homepage: "https://www.washingtonpost.com/", Locale: "en-US", Region: "North America",
			Extractor: source.ExtractorSpec{Type: source.ExtractorFeed, FeedURL: "https://feeds.washingtonpost.com/rss/world"},
		},
		{
			Name: "africa-intelligence", Title: "Africa Intelligence", Homepage: "https://www.africaintelligence.com/", Locale: "en", Region: "Africa",
			Notes: "No public RSS feed advertised; run `news learn https://www.africaintelligence.com/` to derive selectors.",
		},

		// --- pippinlee/news-cli US/UK majors not already covered ---
		{
			Name: "guardian", Title: "The Guardian", Homepage: "https://www.theguardian.com/international", Locale: "en-GB", Region: "Global",
			Extractor: source.ExtractorSpec{Type: source.ExtractorFeed, FeedURL: "https://www.theguardian.com/international/rss"},
		},
		{
			Name: "telegraph", Title: "The Telegraph", Homepage: "https://www.telegraph.co.uk/", Locale: "en-GB", Region: "Europe",
			Extractor: source.ExtractorSpec{Type: source.ExtractorFeed, FeedURL: "https://www.telegraph.co.uk/rss.xml"},
		},
		{
			Name: "independent", Title: "The Independent", Homepage: "https://www.independent.co.uk/", Locale: "en-GB", Region: "Europe",
			Extractor: source.ExtractorSpec{Type: source.ExtractorFeed, FeedURL: "https://www.independent.co.uk/news/rss"},
		},
		{
			Name: "latimes", Title: "Los Angeles Times", Homepage: "https://www.latimes.com/", Locale: "en-US", Region: "North America",
			Extractor: source.ExtractorSpec{Type: source.ExtractorFeed, FeedURL: "https://www.latimes.com/local/rss2.0.xml"},
		},
		{
			Name: "sfchronicle", Title: "San Francisco Chronicle", Homepage: "https://www.sfchronicle.com/", Locale: "en-US", Region: "North America",
			Extractor: source.ExtractorSpec{Type: source.ExtractorFeed, FeedURL: "https://www.sfchronicle.com/rss/feed/Bay-Area-News-429.php"},
		},
		{
			Name: "boston-globe", Title: "The Boston Globe", Homepage: "https://www.bostonglobe.com/", Locale: "en-US", Region: "North America",
			Notes: "No public RSS feed advertised; run `news learn https://www.bostonglobe.com/` to derive selectors.",
		},
		{
			Name: "globe-and-mail", Title: "The Globe and Mail", Homepage: "https://www.theglobeandmail.com/", Locale: "en-CA", Region: "North America",
			Extractor: source.ExtractorSpec{Type: source.ExtractorFeed, FeedURL: "https://www.theglobeandmail.com/arc/outboundfeeds/rss/?outputType=xml"},
		},
		{
			Name: "cbc", Title: "CBC News", Homepage: "https://www.cbc.ca/news", Locale: "en-CA", Region: "North America",
			Extractor: source.ExtractorSpec{Type: source.ExtractorFeed, FeedURL: "https://www.cbc.ca/cmlink/rss-topstories"},
		},
	}
}
