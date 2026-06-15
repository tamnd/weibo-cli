# weibo

[![CI](https://github.com/tamnd/weibo-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/tamnd/weibo-cli/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/tamnd/weibo-cli)](https://github.com/tamnd/weibo-cli/releases/latest)
[![Go Reference](https://pkg.go.dev/badge/github.com/tamnd/weibo-cli.svg)](https://pkg.go.dev/github.com/tamnd/weibo-cli)
[![Go Report Card](https://goreportcard.com/badge/github.com/tamnd/weibo-cli)](https://goreportcard.com/report/github.com/tamnd/weibo-cli)
[![License](https://img.shields.io/github/license/tamnd/weibo-cli)](./LICENSE)

A command line for [weibo.com](https://weibo.com). `weibo` reads hot search
topics, individual posts, comment threads, user profiles, and timelines and
delivers them as clean structured records. One pure-Go binary, no API key.

[Install](#install) • [Commands](#commands) • [Usage](#usage) • [How it works](#how-it-works)

It talks to the public Weibo JSON API over plain HTTPS: dual-host routing
(weibo.com for hot search and suggest, m.weibo.cn for posts and profiles),
mobile browser headers, HTML stripping from post text, and automatic timestamp
normalization to UTC are all handled for you. A session cookie is optional —
pass `--cookie` and `weibo` reaches the gated user and timeline surfaces.

`weibo` is an independent tool. It is not affiliated with Weibo or Sina.

## Install

```bash
go install github.com/tamnd/weibo-cli/cmd/weibo@latest
```

Or grab a prebuilt binary from the [releases](https://github.com/tamnd/weibo-cli/releases):

```bash
# macOS (Homebrew)
brew install tamnd/tap/weibo

# Windows (Scoop)
scoop bucket add tamnd https://github.com/tamnd/scoop-bucket
scoop install weibo

# Debian / Ubuntu
sudo apt install ./weibo_*.deb

# Alpine
apk add --allow-untrusted weibo_*.apk
```

Or run the container image:

```bash
docker run --rm ghcr.io/tamnd/weibo:0.1.1 --help
```

## Commands

| Command | Reads |
| --- | --- |
| `weibo hot` | Weibo hot search board (微博热搜榜), ranked in real time |
| `weibo status <id\|url>` | One Weibo post by numeric id or URL |
| `weibo comments <id\|url>` | Hot comments under a post |
| `weibo suggest <query>` | Search autocomplete suggestions |
| `weibo user <uid\|url>` | A user's public profile (requires `--cookie`) |
| `weibo posts <uid\|url>` | A user's post timeline (requires `--cookie`) |
| `weibo version` | Print version, commit, and build date |

Full reference and guides live at [weibo-cli.tamnd.com](https://weibo-cli.tamnd.com).

## Usage

```bash
# Hot search board — top 30 trending topics right now
weibo hot

# One post — strip HTML, normalize timestamp to UTC
weibo status 5309997458393240

# Comments under a post
weibo comments 5309997458393240 -n 20

# Search autocomplete
weibo suggest 山姆

# User profile — paste your browser SUB cookie
weibo user 2656274875 --cookie 'SUB=xxx; SUBP=yyy'

# User timeline — page 2
weibo posts 2656274875 --page 2 --cookie 'SUB=xxx; SUBP=yyy'

# Or set the cookie once in the environment
export WEIBO_COOKIE='SUB=xxx; SUBP=yyy'
weibo user 2656274875
weibo posts 2656274875 -n 20
```

Records come out as a table (the default on a terminal), JSON, JSONL, CSV,
TSV, or URL list. Table output on a true-color terminal uses aligned columns:

```bash
weibo hot --fields rank,word,heat,url -o table
weibo hot -o jsonl | jq '{rank, word, heat}'
weibo hot -o url

weibo status 5309997458393240 -o json
weibo comments 5309997458393240 -o jsonl > comments.jsonl
```

### Getting a session cookie

Open Weibo in Chrome, open DevTools (F12), go to Application → Cookies →
weibo.com, and copy the `SUB` and `SUBP` values:

```bash
export WEIBO_COOKIE='SUB=<value>; SUBP=<value>'
```

The cookie lives in your browser session. When it expires, copy fresh values.
Keep it out of shell history by using the environment variable.

### Global flags

```
-o, --output     table|list|markdown|json|jsonl|csv|tsv|url|raw
                 (auto: table on TTY, jsonl when piped)
    --fields     comma-separated columns to keep, in order
    --no-header  omit the header row
    --template   Go text/template applied per record
-n, --limit      max records (command default when 0)
    --cookie     session cookie for gated surfaces ("SUB=xxx; SUBP=yyy")
    --user-agent override the User-Agent header
    --rate       min delay between requests (default 500ms)
    --timeout    per-request timeout (default 30s)
    --retries    retry attempts on 429/5xx (default 3)
-q, --quiet      suppress progress output on stderr
-v, --verbose    increase verbosity (repeatable)
    --color      auto|always|never
    --dry-run    print actions without performing them
    --no-cache   bypass on-disk caches
    --data-dir   override the data directory
    --db         tee every record into a store (e.g. out.db)
    --profile    named profile to load
```

## How it works

Weibo's public API is split across two hosts. `weibo` handles that routing for
you:

**Two hosts.** Hot search and suggest live on `weibo.com` and need desktop
Chrome headers. Post detail, comments, user profiles, and timelines live on
`m.weibo.cn` and need mobile Safari headers plus `MWeibo-Pwa: 1` and
`X-Requested-With: XMLHttpRequest`.

**The envelope.** Responses arrive as `{"ok": 1, "data": {...}}`. A value of
`1` means success; `-100` means the surface requires a logged-in session.
`weibo` maps those codes to clear exit statuses.

**HTML in post text.** Post text contains `<a>`, `<span>`, and `<br />`
tags. `weibo` strips them and collapses whitespace so the text field is plain
readable text.

**Timestamps.** Weibo timestamps arrive as `"Mon Jun 15 09:05:12 +0800 2026"`.
`weibo` normalizes them to `"2006-01-02 15:04:05"` UTC.

**What needs a cookie.** `user` and `posts` call
`m.weibo.cn/api/container/getIndex`, which returns ok:-100 for anonymous
callers. Paste your browser `SUB` cookie with `--cookie` or `WEIBO_COOKIE`.
The four other commands work without any cookie.

## Exit codes

```
0  success
2  usage error
3  no results
4  surface requires a login (pass --cookie)
5  rate limited
6  not found
8  network error
```

## Development

```
cmd/weibo/   thin entry point
cli/         kit command assembly and global flags
weibo/       HTTP client, API methods, data models, kit domain registration
docs/        documentation site (Hugo, tago-doks theme)
```

```bash
make build   # ./bin/weibo
make test    # go test ./...
make vet     # go vet ./...
```

Requires Go 1.26+.

## Releasing

Push a version tag and GitHub Actions runs GoReleaser:

```bash
git tag -a v0.1.2 -m "v0.1.2"
git push --tags
```

The image tag carries no `v` prefix (`ghcr.io/tamnd/weibo:0.1.2`).

## License

Apache-2.0. See [LICENSE](LICENSE).

`weibo` is an independent client. Use it to access public data responsibly and
within Weibo's terms of service.
