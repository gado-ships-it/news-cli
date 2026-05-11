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
