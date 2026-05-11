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

		// --- Top 5 responsible sources per continent ---
		//
		// Selection criteria: established editorial standards, broad
		// reach, multilingual diversity. Explicitly excluded per the
		// project's policy: Russian state-controlled outlets (RT, TASS,
		// Sputnik, Channel One, Rossiya, NTV…), North Korean media
		// (KCNA, Rodong Sinmun), and media from any country currently
		// under comprehensive US sanctions (Cuba, Iran, Syria, North
		// Korea, plus regional sanctions affecting Venezuela and
		// Myanmar). Also skipped on editorial-independence grounds:
		// SCMP (post-2020 Hong Kong NSL pressure), Saudi/UAE
		// state-affiliated outlets, Turkish pro-government press.

		// --- Africa ---
		{
			Name: "mg", Title: "Mail & Guardian", Homepage: "https://mg.co.za/", Locale: "en-ZA", Region: "Africa",
			Extractor: source.ExtractorSpec{Type: source.ExtractorFeed, FeedURL: "https://mg.co.za/feed/"},
		},
		{
			Name: "daily-maverick", Title: "Daily Maverick", Homepage: "https://www.dailymaverick.co.za/", Locale: "en-ZA", Region: "Africa",
			Extractor: source.ExtractorSpec{Type: source.ExtractorFeed, FeedURL: "https://www.dailymaverick.co.za/dmrss/"},
		},
		{
			Name: "premium-times", Title: "Premium Times", Homepage: "https://www.premiumtimesng.com/", Locale: "en-NG", Region: "Africa",
			Extractor: source.ExtractorSpec{Type: source.ExtractorFeed, FeedURL: "https://www.premiumtimesng.com/feed"},
		},
		{
			Name: "nation-africa", Title: "Nation Africa (Daily Nation)", Homepage: "https://nation.africa/kenya", Locale: "en-KE", Region: "Africa",
			Notes: "Nation Media Group's public feeds are inconsistent; run `news learn https://nation.africa/kenya` if the bundled extractor stops working.",
		},
		{
			Name: "allafrica", Title: "AllAfrica", Homepage: "https://allafrica.com/", Locale: "en", Region: "Africa",
			Extractor: source.ExtractorSpec{Type: source.ExtractorFeed, FeedURL: "https://allafrica.com/tools/headlines/rdf/latest/headlines.rdf"},
		},

		// --- Middle East / Arab world ---
		{
			Name: "aljazeera", Title: "Al Jazeera English", Homepage: "https://www.aljazeera.com/", Locale: "en", Region: "Middle East",
			Extractor: source.ExtractorSpec{Type: source.ExtractorFeed, FeedURL: "https://www.aljazeera.com/xml/rss/all.xml"},
		},
		{
			Name: "lorient-today", Title: "L'Orient Today", Homepage: "https://today.lorientlejour.com/", Locale: "en-LB", Region: "Middle East",
			Notes: "Beirut-based English daily; feed URL varies — run `news learn` if the bundled extractor fails.",
		},
		{
			Name: "mada-masr", Title: "Mada Masr", Homepage: "https://www.madamasr.com/en/", Locale: "en-EG", Region: "Middle East",
			Notes: "Mada Masr's RSS endpoint blocks non-browser User-Agents; run `news learn https://www.madamasr.com/en/` to derive a homepage extractor.",
		},
		{
			Name: "haaretz", Title: "Haaretz", Homepage: "https://www.haaretz.com/", Locale: "en-IL", Region: "Middle East",
			Notes: "Most Haaretz feeds sit behind the subscriber wall; run `news learn https://www.haaretz.com/` to derive a homepage extractor.",
		},
		{
			Name: "al-monitor", Title: "Al-Monitor", Homepage: "https://www.al-monitor.com/", Locale: "en", Region: "Middle East",
			Extractor: source.ExtractorSpec{Type: source.ExtractorFeed, FeedURL: "https://www.al-monitor.com/rss"},
		},

		// --- Asia ---
		{
			Name: "the-hindu", Title: "The Hindu", Homepage: "https://www.thehindu.com/", Locale: "en-IN", Region: "Asia",
			Extractor: source.ExtractorSpec{Type: source.ExtractorFeed, FeedURL: "https://www.thehindu.com/news/feeder/default.rss"},
		},
		{
			Name: "nikkei-asia", Title: "Nikkei Asia", Homepage: "https://asia.nikkei.com/", Locale: "en", Region: "Asia",
			Extractor: source.ExtractorSpec{Type: source.ExtractorFeed, FeedURL: "https://asia.nikkei.com/rss/feed/nar"},
		},
		{
			Name: "asahi", Title: "The Asahi Shimbun", Homepage: "https://www.asahi.com/ajw/", Locale: "en-JP", Region: "Asia",
			Notes: "Asahi's English (AJW) RSS endpoints rotate; run `news learn https://www.asahi.com/ajw/` if the bundled extractor fails.",
		},
		{
			Name: "dawn", Title: "Dawn", Homepage: "https://www.dawn.com/", Locale: "en-PK", Region: "Asia",
			Extractor: source.ExtractorSpec{Type: source.ExtractorFeed, FeedURL: "https://www.dawn.com/feeds/home"},
		},
		{
			Name: "straits-times", Title: "The Straits Times", Homepage: "https://www.straitstimes.com/", Locale: "en-SG", Region: "Asia",
			Notes: "Straits Times serves feeds only behind a section URL; run `news learn https://www.straitstimes.com/` for the firehose.",
		},

		// --- Europe (broadening beyond UK/Switzerland already in seed) ---
		{
			Name: "faz", Title: "Frankfurter Allgemeine Zeitung", Homepage: "https://www.faz.net/", Locale: "de-DE", Region: "Europe",
			Extractor: source.ExtractorSpec{Type: source.ExtractorFeed, FeedURL: "https://www.faz.net/rss/aktuell/"},
		},
		{
			Name: "sz", Title: "Süddeutsche Zeitung", Homepage: "https://www.sueddeutsche.de/", Locale: "de-DE", Region: "Europe",
			Extractor: source.ExtractorSpec{Type: source.ExtractorFeed, FeedURL: "https://rss.sueddeutsche.de/rss/Topthemen"},
		},
		{
			Name: "spiegel", Title: "Der Spiegel", Homepage: "https://www.spiegel.de/", Locale: "de-DE", Region: "Europe",
			Extractor: source.ExtractorSpec{Type: source.ExtractorFeed, FeedURL: "https://www.spiegel.de/schlagzeilen/index.rss"},
		},
		{
			Name: "elpais", Title: "El País", Homepage: "https://elpais.com/", Locale: "es-ES", Region: "Europe",
			Extractor: source.ExtractorSpec{Type: source.ExtractorFeed, FeedURL: "https://feeds.elpais.com/mrss-s/pages/ep/site/elpais.com/portada"},
		},
		{
			Name: "politico-eu", Title: "Politico Europe", Homepage: "https://www.politico.eu/", Locale: "en", Region: "Europe",
			Extractor: source.ExtractorSpec{Type: source.ExtractorFeed, FeedURL: "https://www.politico.eu/feed/"},
		},

		// --- North America (newswires + investigative + long-form) ---
		{
			Name: "npr", Title: "NPR", Homepage: "https://www.npr.org/", Locale: "en-US", Region: "North America",
			Extractor: source.ExtractorSpec{Type: source.ExtractorFeed, FeedURL: "https://feeds.npr.org/1001/rss.xml"},
		},
		{
			Name: "propublica", Title: "ProPublica", Homepage: "https://www.propublica.org/", Locale: "en-US", Region: "North America",
			Extractor: source.ExtractorSpec{Type: source.ExtractorFeed, FeedURL: "https://www.propublica.org/feeds/propublica/main"},
		},
		{
			Name: "atlantic", Title: "The Atlantic", Homepage: "https://www.theatlantic.com/", Locale: "en-US", Region: "North America",
			Extractor: source.ExtractorSpec{Type: source.ExtractorFeed, FeedURL: "https://www.theatlantic.com/feed/all/"},
		},
		{
			Name: "bloomberg", Title: "Bloomberg", Homepage: "https://www.bloomberg.com/", Locale: "en-US", Region: "North America",
			Notes: "Bloomberg restricts public feeds to a small set of topics; run `news learn https://www.bloomberg.com/` to derive a wider extractor.",
		},
		{
			Name: "ap", Title: "Associated Press", Homepage: "https://apnews.com/", Locale: "en-US", Region: "North America",
			Notes: "AP retired most public RSS endpoints; run `news learn https://apnews.com/` to derive a homepage extractor.",
		},

		// --- Oceania ---
		{
			Name: "abc-au", Title: "ABC News (Australia)", Homepage: "https://www.abc.net.au/news", Locale: "en-AU", Region: "Oceania",
			Extractor: source.ExtractorSpec{Type: source.ExtractorFeed, FeedURL: "https://www.abc.net.au/news/feed/45910/rss.xml"},
		},
		{
			Name: "smh", Title: "The Sydney Morning Herald", Homepage: "https://www.smh.com.au/", Locale: "en-AU", Region: "Oceania",
			Extractor: source.ExtractorSpec{Type: source.ExtractorFeed, FeedURL: "https://www.smh.com.au/rss/feed.xml"},
		},
		{
			Name: "the-age", Title: "The Age", Homepage: "https://www.theage.com.au/", Locale: "en-AU", Region: "Oceania",
			Extractor: source.ExtractorSpec{Type: source.ExtractorFeed, FeedURL: "https://www.theage.com.au/rss/feed.xml"},
		},
		{
			Name: "rnz", Title: "Radio New Zealand", Homepage: "https://www.rnz.co.nz/", Locale: "en-NZ", Region: "Oceania",
			Extractor: source.ExtractorSpec{Type: source.ExtractorFeed, FeedURL: "https://www.rnz.co.nz/rss/national.xml"},
		},
		{
			Name: "nzherald", Title: "The New Zealand Herald", Homepage: "https://www.nzherald.co.nz/", Locale: "en-NZ", Region: "Oceania",
			Notes: "NZ Herald's public RSS endpoints are returning 5xx; run `news learn https://www.nzherald.co.nz/` to derive a homepage extractor.",
		},

		// --- South America ---
		{
			Name: "folha", Title: "Folha de S.Paulo", Homepage: "https://www1.folha.uol.com.br/", Locale: "pt-BR", Region: "South America",
			Extractor: source.ExtractorSpec{Type: source.ExtractorFeed, FeedURL: "https://feeds.folha.uol.com.br/emcimadahora/rss091.xml"},
		},
		{
			Name: "oglobo", Title: "O Globo", Homepage: "https://oglobo.globo.com/", Locale: "pt-BR", Region: "South America",
			Extractor: source.ExtractorSpec{Type: source.ExtractorFeed, FeedURL: "https://oglobo.globo.com/rss.xml"},
		},
		{
			Name: "la-nacion", Title: "La Nación", Homepage: "https://www.lanacion.com.ar/", Locale: "es-AR", Region: "South America",
			Notes: "La Nación's feed paths rotate; run `news learn https://www.lanacion.com.ar/` to derive a stable extractor.",
		},
		{
			Name: "clarin", Title: "Clarín", Homepage: "https://www.clarin.com/", Locale: "es-AR", Region: "South America",
			Extractor: source.ExtractorSpec{Type: source.ExtractorFeed, FeedURL: "https://www.clarin.com/rss/lo-ultimo/"},
		},
		{
			Name: "el-espectador", Title: "El Espectador", Homepage: "https://www.elespectador.com/", Locale: "es-CO", Region: "South America",
			Notes: "El Espectador serves only paywalled per-section feeds; run `news learn https://www.elespectador.com/` for a homepage extractor.",
		},
	}
}
