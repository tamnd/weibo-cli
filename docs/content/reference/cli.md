---
title: "CLI"
description: "Every command and flag, with their defaults."
weight: 10
---

Run `weibo <command> --help` for the live flag list on any command. Every
command accepts the [global flags](#global-flags) and renders through the
shared [output formatter](/reference/output/).

## Commands

### Open commands (no cookie needed)

| Command | Argument | What it does |
|---|---|---|
| `hot` | — | Weibo hot search board, ranked in real time; default limit 30 |
| `status <id>` | numeric id or post URL | One Weibo post; HTML stripped, timestamp UTC |
| `comments <id>` | numeric id or post URL | Hot comments under a post; default limit 20 |
| `suggest <query>` | search terms | Search autocomplete suggestions; default limit 10 |

**Accepted id forms for `status` and `comments`:**
- Bare numeric id: `5309997458393240`
- `https://m.weibo.cn/detail/5309997458393240`
- `https://m.weibo.cn/status/5309997458393240`

### Gated commands (require `--cookie`)

| Command | Argument | What it does |
|---|---|---|
| `user <uid>` | numeric uid or profile URL | Public user profile |
| `posts <uid>` | numeric uid or profile URL | User's post timeline; `--page` for pagination |

**Accepted uid forms for `user` and `posts`:**
- Bare numeric uid: `2656274875`
- `https://weibo.com/u/2656274875`
- `https://m.weibo.cn/u/2656274875`
- `https://m.weibo.cn/profile/2656274875`

Without `--cookie` or `WEIBO_COOKIE`, both commands exit 4 immediately.

### Utility

| Command | What it does |
|---|---|
| `version` | Print version, commit, and build date |

## Command flags

### `posts` flags

| Flag | Default | Meaning |
|---|---|---|
| `--page` | `1` | Page number (1-based); Weibo serves ~10 posts per page |
| `-n, --limit` | `10` | Max posts to return from that page |

## Global flags

Available on every command:

| Flag | Default | Meaning |
|---|---|---|
| `-o, --output` | auto | `table`, `list`, `markdown`, `json`, `jsonl`, `csv`, `tsv`, `url`, `raw` |
| `-n, --limit` | command default | Max records; `0` means use the command default |
| `--fields` | all | Comma-separated columns to keep, in order |
| `--template` | none | Go `text/template` applied per record |
| `--no-header` | off | Omit the header row in table/csv/tsv output |
| `--cookie` | none | Session cookie for gated commands (`"SUB=xxx; SUBP=yyy"`) |
| `--user-agent` | Chrome/Safari | Override the User-Agent header |
| `--rate` | `500ms` | Minimum delay between requests |
| `--timeout` | `30s` | Per-request timeout |
| `--retries` | `3` | Retry attempts on 429 or 5xx |
| `-q, --quiet` | off | Suppress progress output on stderr |
| `-v, --verbose` | `0` | Increase verbosity (repeatable: `-vv`) |
| `--color` | `auto` | `auto`, `always`, or `never` |
| `--dry-run` | off | Print actions without performing them |
| `--no-cache` | off | Bypass on-disk caches |
| `--data-dir` | `~/.local/share/weibo` | Override the data directory |
| `--db` | none | Tee every record into a store (e.g. `out.db`, `postgres://...`) |
| `--profile` | none | Named profile to load |

The output default adapts to where it is going: an aligned table when writing
to a terminal, JSONL when piped. See [output formats](/reference/output/).

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Success |
| `2` | Usage error (bad flag or argument) |
| `3` | No results found |
| `4` | Surface requires a login — pass `--cookie` |
| `5` | Rate limited by Weibo |
| `6` | Not found |
| `8` | Network error |
