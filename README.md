<img width="1254" height="1254" style="width:200px;" alt="news-cli" src="news-cli.png" />

# news-cli

**An agent-native CLI for news — inspired by [tenor.news](https://tenor.news/).**

Independent project. `tenor.news` is the original user-facing service
that monitors global outlets, filters out clickbait, removes duplicates
and condenses the day's events; `news-cli` borrows that editorial posture
and brings it to the terminal so an LLM agent on your machine can do the
same job locally — against whatever lineup of outlets you choose, with
formatting an agent actually wants.

It fetches headlines, deks (subtitles/abstracts) and lede images from
publicly available frontpages and emits them in three formats: terse text
(default), Markdown (`--md`), or JSON (`--json`). The JSON output is sized
to fit comfortably in an LLM context window so an agent can summarize the
current state of the world to your taste.

## What it does

- **(a) Frontpages as data.** Headlines, deks, links, lede images — never article bodies.
- **(b) LLM-learned extractors.** `news learn <url>` discovers an outlet's RSS / Atom feed, falls back to schema.org `NewsArticle` in `ld+json`, and finally shells out to your locally-installed coding-agent CLI (`claude` or `codex`) to derive CSS selectors against the homepage HTML. No API key handling — auth comes from whichever CLI you already have.
- **(c) CLI-first.** `--md` and `--json` work on every subcommand.
- **(d) Community sources.** When `learn` succeeds you can submit the new source back to this repo with `news submit <name>`.
- **(+) Editorial brief.** `news tenor` asks the LLM to dedupe, drop clickbait, and surface long-lasting stories — optionally biased to your interest of the day.

## Install

```sh
go install github.com/gado-ships-it/news-cli@latest
```

Or from the repo:

```sh
git clone https://github.com/gado-ships-it/news-cli && cd news-cli
go build -o news .
```

## Usage at a glance

```sh
news fetch                                    # every configured source
news fetch nzz nyt bbc                        # a subset
news fetch --md                               # markdown
news fetch --json -n 5                        # JSON, 5 items per source
news fetch --ascii --ascii-width 60           # lede images as ASCII art

news list                                     # see what's available
news list --json

news tenor                                    # deduped editorial brief
news tenor -i "climate science only"          # bias to a topic
news tenor --md --ascii --entries 10          # bigger, with thumbnails

news learn https://www.example-news.com/
news submit example                           # opens a PR via `gh`
```

## Examples

Real outputs from `news fetch` and `news tenor`. Headlines and excerpts are condensed for readability; URLs are shown verbatim.

### Plain text — `news fetch`

```text
$ news fetch bbc nzz -n 2
=== BBC News — https://www.bbc.com/news ===
  • Analysis: Has Starmer done enough to save his premiership?
    Was the prime minister's speech enough to avert a challenge to his leadership less than two years after he won a landslide general election victory?
    https://www.bbc.com/news/articles/c3r2pr95yq0o

  • British passengers from cruise ship isolating in hospital
    The passengers landed in the UK on Sunday and none have reported symptoms, but they will be monitored in hospital for 72 hours.
    https://www.bbc.com/news/articles/c4g83vddnz0o

=== Neue Zürcher Zeitung — https://www.nzz.ch/ ===
  • Zwei chinesische Stromer für Stellantis: Leapmotor B05 und B03X sollen dem Europa-Partner neue Wege aufzeigen
    Mit leistungsstarker Technik zu tiefen Preisen stiehlt Leapmotor den Fahrzeugen des europäischen Joint-Venture-Partners die Schau.
    https://www.nzz.ch/mobilitaet/zwei-chinesische-stromer-fuer-stellantis-ld.10006356

  • Der Schattenmann der thailändischen Politik darf in den Hausarrest
    Am Montag wurde der ehemalige thailändische Ministerpräsident Thaksin Shinawatra aus der Haft entlassen.
    https://www.nzz.ch/international/der-schattenmann-der-thailaendischen-politik-darf-ld.10006603
```

On a TTY, headlines render in **bold bright cyan**, source banners in **bold bright yellow**, and URLs as gray + underlined [OSC 8](https://gist.github.com/egmontkob/eb114294efbcd5adb1944c9f3cb5feda) hyperlinks — click them directly in your terminal.

### Markdown — `news fetch --md`

```text
$ news fetch nzz -n 1 --md
# News frontpages — fetched 2026-05-11T15:51:13Z

## Neue Zürcher Zeitung

Source: [https://www.nzz.ch/](https://www.nzz.ch/) · fetched 2026-05-11T15:51:13Z

- **[Zwei chinesische Stromer für Stellantis: Leapmotor B05 und B03X sollen dem Europa-Partner neue Wege aufzeigen](https://www.nzz.ch/mobilitaet/zwei-chinesische-stromer-fuer-stellantis-ld.10006356)** — _via Neue Zürcher Zeitung_
  > Mit leistungsstarker Technik zu tiefen Preisen stiehlt Leapmotor den Fahrzeugen des europäischen Joint-Venture-Partners die Schau. Zudem soll der Hersteller bald ein neues Auto für Opel entwickeln.
```

### JSON — `news fetch --json`

Designed for an LLM agent or downstream script. Every item carries `source`, `source_url`, the article `url`, and a `fetched_at` timestamp so citations stay intact when piped into another tool.

```json
$ news fetch bbc -n 1 --json
[
  {
    "source": {
      "name": "bbc",
      "title": "BBC News",
      "homepage": "https://www.bbc.com/news",
      "locale": "en-GB",
      "region": "Global",
      "extractor": {
        "type": "feed",
        "feed_url": "https://feeds.bbci.co.uk/news/rss.xml"
      }
    },
    "items": [
      {
        "source": "bbc",
        "source_url": "https://www.bbc.com/news",
        "headline": "Analysis: Has Starmer done enough to save his premiership?",
        "dek": "Was the prime minister's speech enough to avert a challenge to his leadership less than two years after he won a landslide general election victory?",
        "url": "https://www.bbc.com/news/articles/c3r2pr95yq0o",
        "published": "2026-05-11T13:09:32Z",
        "fetched_at": "2026-05-11T15:51:13Z"
      }
    ],
    "fetched_at": "2026-05-11T15:51:13Z"
  }
]
```

### Editorial brief — `news tenor`

Fetches every configured source (or a subset), ships the corpus to your locally-installed `claude` or `codex` CLI, and returns a deduplicated, long-lasting-focused brief. The example below merged stories across BBC, NYT and Le Monde in ~13 seconds.

```text
$ news tenor bbc nyt lemonde -n 5 --entries 3
fetching 3 sources…
asking claude for a brief over 15 headlines from 3 sources…
News brief — 2026-05-11  (via claude)

• Trump-Xi summit looms as China prepares for trade confrontation
  Ahead of President Trump's visit to Beijing, China is signaling readiness for an economic showdown
  and building a legal arsenal, while US lawmakers press the administration to advance a delayed
  arms sale to Taiwan. Asian middle powers fear Washington may trade security commitments for better
  economic terms with Beijing.
  sources: nyt, nyt, nyt
    https://www.nytimes.com/2026/05/11/business/trump-xi-economic-warfare.html
    https://www.nytimes.com/2026/05/11/us/politics/taiwan-trump-china-xi-jinping.html
    https://www.nytimes.com/2026/05/11/world/asia/trump-xi-china-summit-iran.html

• Middle East war drives major energy shock and global economic strain
  Aramco's CEO described the Iran conflict as the largest energy shock on record, warning that
  markets would need months to rebalance even if the Strait of Hormuz reopened immediately. India's
  Prime Minister Modi is urging citizens to curb gold purchases and foreign travel.
  sources: lemonde, nyt
    https://www.lemonde.fr/international/live/2026/05/11/en-direct-guerre-au-moyen-orient-...
    https://www.nytimes.com/2026/05/11/world/asia/modi-indians-gold-weddings-fuel-economy-iran.html

• Hantavirus outbreak on cruise ship triggers international repatriations
  Passengers from an affected cruise ship are being returned to their home countries, with British
  arrivals isolated in hospital for 72-hour monitoring and a repatriated French citizen testing
  positive. The UN has called for coordinated international health efforts.
  sources: bbc, lemonde, nyt
    https://www.bbc.com/news/articles/c4g83vddnz0o
    https://www.lemonde.fr/sante/live/2026/05/11/en-direct-hantavirus-le-hondius-...
    https://www.nytimes.com/2026/05/11/podcasts/the-headlines/trump-iran-hantavirus.html
```

On a TTY the URL list collapses to **`sources: bbc, lemonde, nyt`** where each name is a clickable hyperlink (gray + underlined).

### Topic-biased brief — `news tenor -i "…"`

Free-form prose biases the brief. Narrows ("only X"), expands ("more X, less Y"), or overrides defaults ("yes I want sports, that's exactly what I want") — whichever the model judges from the prose.

```text
$ news tenor -i "today I only want science and climate, no politics" \
            bbc nyt nzz lemonde economist nature --entries 5
News brief — 2026-05-11  (via claude)

• Hantavirus outbreak from cruise ship 'Hondius' spreads to multiple countries
  Passengers from the affected cruise are being repatriated and monitored across Europe and beyond,
  with new positive cases in the UK and France and three deaths so far linked to the outbreak.
  sources: bbc, bbc, nzz, lemonde

• Why hantavirus is unlikely to trigger a Covid-style pandemic
  An NZZ science analysis lays out four reasons the current hantavirus cluster differs fundamentally
  from SARS-CoV-2, including its transmission route and lack of efficient human-to-human spread.
  sources: nzz

• The sleep paradox: why humans sleep less than biology suggests we need
  Nature explores emerging research on why human sleep duration appears mismatched with its
  restorative importance.
  sources: nature

• Drug developers make progress against 'undruggable' cancer proteins
  Nature reports on new chemical approaches that are beginning to target cancer-driving proteins
  long considered inaccessible to small-molecule drugs.
  sources: nature

• Presymptomatic training shown to ease deficits in Rett syndrome mouse model
  A publisher correction accompanies findings that early behavioural training in a mouse model
  of Rett syndrome can mitigate later functional impairments.
  sources: nature
```

Trump-China and Iran politics dominated the same corpus when run without `-i`; here they're absent.

### ASCII art — `--ascii`

Pre-rendered concurrently from each item's lede image. PNG/JPEG/GIF supported via stdlib. Brightness is mapped to a 10-step glyph ramp (` .:-=+*#%@`), heights are aspect-corrected for terminal cell ratio.

```text
$ news fetch nyt -n 1 --ascii --ascii-width 50
=== The New York Times — https://www.nytimes.com/ ===
  • As Trump Heads to Beijing, China Is 'Locked and Loaded' for a Fight
    Beijing is signaling that it is ready for a trade showdown, and it is building up a legal arsenal in preparation.
    https://www.nytimes.com/2026/05/11/business/trump-xi-economic-warfare.html
    ####********#*****#*#*##*#######*######********+++
    ####*#*********************************##*********
    ##########***#########****++**++++++++++++++++++++
    #################*###*######**+++++***************
    *#*****************##****#**####******++++++******
    *******++*+****+++++**++*************+++++++++++++
    ***************************+**+*+++**********+++++
    ****#*#************++*********++++*****+++++++++++
    *************++*+*++++++++++++++++++++*****++++*++
    **************+*++++++++++++++++++++++++++++++++++
    ********#%#*******++++*+++++++++++++++++++++++++++
    *******=*****+*+*+++++++++++++++++++++++++++++++++
    ********++++++++++++++++++++++++++++++++++++++++++
    *=---==+:-...:...:::--##=*-%===--==*--------=+====
    =:=+++++++.....-::...::#:.-::.:-=-.==*-.+..-:..-=:
    -==:.:.::.::::.:.-=.+-::-.:::...-.*..-::+...:---:.
    .==-.:.::=:-:.. -- ::-.-. :- ..:...  .:.:: -:.. .-
    ====::::.--.==:.--. .:::.=---:-=:=*---.--:--...::.
    .::::.=-:..:-= -::--::.-=  : ...::. .  ..:--...+:.
    ...-.. :::.=- ..::.:.-==.*--=..+==--=:..:...:.+:#:
    ...::  --.::..  .*=+ ===   -+ .  :.   ::.=--= ..#.
    =. :.. -****+=++.. . *+**++=+ %---- .=   ..: :..=:
    === ..+--:--:==-- .  .*-=+--+..     :    .:     ::
    -:- +::.::::: =.==. :  : - ..   :   :::.: .   .=-:
    --- .    .::=++=.-= -:   - ..=.   ....+--=-::::..:
```

Combine with `--md` to embed the art in fenced code blocks under each headline, or with `news tenor --ascii` to attach one representative thumbnail to each merged brief entry.

### Source list — `news list`

```text
$ news list
abc-au                 ABC News (Australia)              https://www.abc.net.au/news [feed]
africa-intelligence    Africa Intelligence               https://www.africaintelligence.com/ [unconfigured]
african-arguments      African Arguments                 https://africanarguments.org/ [feed]
al-monitor             Al-Monitor                        https://www.al-monitor.com/ [feed]
aljazeera              Al Jazeera English                https://www.aljazeera.com/ [feed]
allafrica              AllAfrica                         https://allafrica.com/ [feed]
ap                     Associated Press                  https://apnews.com/ [unconfigured]
asahi                  The Asahi Shimbun                 https://www.asahi.com/ajw/ [unconfigured]
atlantic               The Atlantic                      https://www.theatlantic.com/ [feed]
bbc                    BBC News                          https://www.bbc.com/news [feed]
...
```

62 sources total — bundled seed list spans Africa, the Middle East, Asia, Europe, North America, Oceania, and South America. Use `news list --json` for machine-readable output.

## Subcommands

| Command | What it does |
|---|---|
| `fetch [name…]` | Fetch one, several, or all configured frontpages. |
| `list` | Show every configured source. |
| `tenor [name…]` | Deduplicated editorial brief via `claude` / `codex`. |
| `learn <url>` | Derive an extractor (feed → ld+json → CSS via LLM) and save it locally. |
| `submit <name>` | Open a PR to `gado-ships-it/news-cli` adding a locally-learned source. |
| `sources show <name>` | Print one source's JSON definition. |
| `sources remove <name>` | Drop a source from your local store. |
| `sources path` | Print the path to `sources.json`. |

### Global flags (every subcommand)

| Flag | Default | Meaning |
|---|---|---|
| `--md` | off | Render output as Markdown. |
| `--json` | off | Render output as JSON. `--json` wins if both are given. |
| `--ascii` | off | Render lede images as ASCII art under each item (text and `--md` only; ignored for `--json`). |
| `--ascii-width N` | `60` | ASCII art width in characters. Height is derived from the source aspect ratio with terminal cell correction. |

When neither `--md` nor `--json` is set, the **default text mode** uses ANSI styling on a TTY: bold cyan headlines, bold yellow source banners, and gray + underlined OSC 8 hyperlinks for URLs and source citations. Styling is auto-disabled when stdout isn't a TTY or `NO_COLOR` is set; `FORCE_COLOR=1` overrides the TTY check.

### `news fetch [name…]`

Fetches the named sources (or every configured source if no names are given) in parallel.

| Flag | Short | Default | Meaning |
|---|---|---|---|
| `--max N` | `-n N` | `10` | Max items per source. `0` = unlimited. |
| `--concurrency N` |  | `6` | Parallel fetches across sources. |
| `--timeout D` |  | `30s` | Overall fetch timeout (Go duration: `30s`, `1m30s`, …). |

### `news tenor [name…]`

Fetches the named sources (or all configured), ships the corpus to your locally-installed LLM CLI, and prints a short editorial brief: duplicates merged, single-event noise dropped, long-term-consequential stories surfaced. Every entry cites at least one source with the verbatim article URL.

| Flag | Short | Default | Meaning |
|---|---|---|---|
| `--max N` | `-n N` | `10` | Max items per source sent to the LLM (caps prompt size). |
| `--entries N` |  | `8` | Target number of merged entries in the brief. |
| `--interest "…"` | `-i` | (empty) | Free-form prose to bias topic selection. Examples: `"politics from Asia"`, `"only sports"`, `"climate and science, no politics"`. When this contradicts the default "long-lasting / non-clickbait" rules, the interest wins. |
| `--concurrency N` |  | `6` | Parallel fetches across sources. |
| `--timeout D` |  | `5m` | Overall fetch + LLM timeout. |

### `news learn <homepage-url>`

Discovery pipeline, in order:

1. **RSS / Atom feed** — `<link rel="alternate">` autodiscovery, anchor scan for feed-shaped URLs (`*.rss`, `*.atom`, `/feeds/`…), one-level follow of `/rss` and `/feeds` index pages, and a fixed list of common paths. Candidates are ranked by name hints (`recent`, `all`, `top`, `headlines`…) so multi-feed outlets land on the firehose, not a section feed.
2. **schema.org `NewsArticle`** — walks `<script type="application/ld+json">` blocks on the homepage for `NewsArticle` / `ItemList` entries.
3. **LLM-derived CSS selectors** — sends a truncated, cleaned copy of the homepage HTML to `claude -p` or `codex exec`, asks for `{item, headline, link, dek, image}` selectors, and validates them against the same page before saving.

| Flag | Default | Meaning |
|---|---|---|
| `--name <id>` | derived from host | Override the auto-derived short id (`nzz.ch` → `nzz`). |
| `--title <text>` | derived from host | Override the auto-derived display title. |
| `--yes` | off | Skip the "submit back via PR?" prompt. |
| `--timeout D` | `1m30s` | Overall learn timeout. |

### `news submit <name>`

Packages a locally-learned source as a contribution to the shared `gado-ships-it/news-cli` source list. If the [GitHub CLI](https://cli.github.com/) (`gh`) is installed and authenticated, the PR is opened in your browser via `gh pr create --web`; otherwise the encoded source is printed with manual fork-and-PR instructions.

### `news sources`

| Subcommand | What it does |
|---|---|
| `sources show <name>` | Print the resolved JSON definition of one source (local store wins over seed). |
| `sources remove <name>` | Remove a learned source from the local store. Seeded sources are unaffected. |
| `sources path` | Print the path to your local `sources.json`. |

## Environment

| Variable | Purpose |
|---|---|
| `NEWS_CLI_LLM` | Force a specific LLM backend: `claude` or `codex`. Default is `claude` if present, otherwise `codex`. |
| `NO_COLOR` | Disable all ANSI styling regardless of TTY status (see [no-color.org](https://no-color.org/)). |
| `FORCE_COLOR` | Force ANSI styling even when stdout isn't a TTY — useful for piping through `less -R`. |

## Configuration

- Local source overrides live at `~/.config/news-cli/sources.json`.
- The bundled seed list is overlaid by your local store (yours wins).
- LLM-driven subcommands (`learn`'s CSS fallback and the whole of `tenor`) require either the [Claude Code CLI](https://docs.anthropic.com/en/docs/claude-code) (`claude`) or the [OpenAI Codex CLI](https://github.com/openai/codex) (`codex`) installed and authenticated on `$PATH`. Auth lives in the CLI — no API key is read from this tool's environment.

## Bundled sources

The starter seed list is the lineup [tenor.news](https://tenor.news/) was
using when this project was written (used here as a sensible default, not
an endorsement), extended with the major US/UK outlets from
[pippinlee/news-cli](https://github.com/pippinlee/news-cli). `news-cli`
is not affiliated with either project — add, remove and override sources
with `news learn`, `news submit`, and `news sources remove` to make the
list your own.

Semafor · Neue Zürcher Zeitung · The New York Times · The Economist · Foreign Policy ·
African Arguments · Le Monde · The Korea Herald · CNA (Channel News Asia) · Nature ·
Tsüri.ch · Le Monde diplomatique · SRF · Republik · BBC · The Washington Post ·
Africa Intelligence · The Guardian · The Telegraph · The Independent · LA Times ·
San Francisco Chronicle · Boston Globe · The Globe and Mail · CBC.

A handful (e.g. Africa Intelligence) ship without a configured extractor —
run `news learn` against their homepage to derive one.

## Legal posture

`news-cli` is built to **drive traffic to publishers**, not replace them.

- Only **publicly accessible** frontpage metadata is fetched (headline + dek + lede image + link). Article bodies are never downloaded.
- The publisher's **homepage and article URL are always present** in every output — text, Markdown, and JSON — so an LLM digesting the output keeps citation links intact.
- The HTTP `User-Agent` identifies the tool and links to this repository so publishers can rate-limit or block it as they prefer.
- No paywall circumvention. If an article is paywalled, the link still goes to the publisher's paywall — that is the point.
- No caching of fetched content to disk by default.

If you are a publisher and you would like your outlet removed from the seed
list, please open an issue at `gado-ships-it/news-cli` and the source will be
unbundled in the next release.

## Adding a source

```sh
# 1. derive an extractor
news learn https://your-favorite-paper.com/

# 2. test it
news fetch <name> -n 5

# 3. share it
news submit <name>
```

`learn` walks three discovery layers in order — feeds, JSON-LD, then LLM
selectors — and refuses to save a source unless the chosen extractor
actually returns items.
