# news-cli

Read public news frontpages from your terminal — designed to be agent-friendly.

`news-cli` fetches headlines, deks (subtitles/abstracts) and lede images from
the publicly available frontpages of news outlets, and emits them in three
formats: terse text (default), Markdown (`--md`), or JSON (`--json`). The
JSON output is sized to fit comfortably in an LLM context window so that an
agent can summarize the current state of the world to your taste.

## What it does

- **(a) Frontpages as data.** Headlines, deks, links, lede images — never article bodies.
- **(b) LLM-learned extractors.** `news learn <url>` discovers an outlet's RSS / Atom feed, falls back to schema.org `NewsArticle` in `ld+json`, and finally shells out to your locally-installed coding-agent CLI (`claude` or `codex`) to derive CSS selectors against the homepage HTML. No API key handling — auth comes from whichever CLI you already have.
- **(c) CLI-first.** `--md` and `--json` work on every subcommand.
- **(d) Community sources.** When `learn` succeeds you can submit the new source back to this repo with `news submit <name>`.

## Install

```sh
go install github.com/gado-ships-it/news-cli@latest
```

Or from the repo:

```sh
git clone https://github.com/gado-ships-it/news-cli && cd news-cli
go build -o news .
```

## Usage

```sh
news fetch                # every configured source
news fetch nzz nyt bbc    # a subset
news fetch --md           # markdown
news fetch --json -n 5    # JSON, 5 items per source

news list                 # see what's available
news list --json

news learn https://www.example-news.com/
news submit example       # opens a PR to gado-ships-it/news-cli via `gh`
```

### Subcommands

| Command | What it does |
|---|---|
| `fetch [name…]` | Fetch one, several, or all configured frontpages. |
| `list` | Show every configured source. |
| `learn <url>` | Derive an extractor (feed → ld+json → CSS via LLM) and save it locally. |
| `submit <name>` | Open a PR to `gado-ships-it/news-cli` adding a locally-learned source. |
| `sources show <name>` | Print one source's JSON definition. |
| `sources remove <name>` | Drop a source from your local store. |
| `sources path` | Print the path to `sources.json`. |

Every subcommand accepts `--md` and `--json`.

### Configuration

- Local source overrides live at `~/.config/news-cli/sources.json`.
- The bundled seed list is overlaid by your local store (yours wins).
- LLM CSS fallback requires either the [Claude Code CLI](https://docs.anthropic.com/en/docs/claude-code) (`claude`) or the [OpenAI Codex CLI](https://github.com/openai/codex) (`codex`) installed and authenticated on `$PATH`. Auth lives in the CLI — no API key is read from this tool's environment. `claude` is preferred when both are available; force one with `NEWS_CLI_LLM=claude` or `NEWS_CLI_LLM=codex`.

## Bundled sources

The seed list combines the outlets used by [tenor.news](https://tenor.news/)
with the major US/UK outlets from [pippinlee/news-cli](https://github.com/pippinlee/news-cli):

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
