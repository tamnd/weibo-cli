---
title: "Configuration"
description: "Environment variables and global flags, with their defaults."
weight: 20
---

`weibo` needs no configuration file. Every option is a flag or an environment
variable, and the defaults are chosen so the four open commands need neither.

## Session cookie

`user` and `posts` read from endpoints that return ok:-100 for anonymous
callers. Supply a session cookie from a logged-in browser:

```bash
# inline flag (stays out of shell history if you quote it)
weibo user 2656274875 --cookie 'SUB=xxx; SUBP=yyy'

# preferred: set it once in the environment
export WEIBO_COOKIE='SUB=xxx; SUBP=yyy'
weibo user 2656274875
weibo posts 2656274875
```

**How to get the cookie:**

1. Open [weibo.com](https://weibo.com) in Chrome and log in.
2. Press F12, go to **Application → Cookies → weibo.com**.
3. Copy `SUB` and `SUBP` values, paste them as `SUB=<value>; SUBP=<value>`.

The cookie is your live browser session. Treat it like a password — do not
commit it to source control or put it in shell history. When it expires (a
few days to a few weeks depending on your session), copy fresh values.

## Environment variables

| Variable | Used for |
|---|---|
| `WEIBO_COOKIE` | Session cookie for `user` and `posts` (`SUB=xxx; SUBP=yyy`) |

## Global flags

| Flag | Default | Meaning |
|---|---|---|
| `-o, --output` | auto | `table`, `json`, `jsonl`, `csv`, `tsv`, `url` |
| `-n, --limit` | command default | Max records per command |
| `--fields` | all | Comma-separated columns to keep, in order |
| `--template` | none | Go `text/template` applied per record |
| `--no-header` | off | Omit the header row in table/csv/tsv output |
| `--cookie` | none | Session cookie for gated commands |
| `--user-agent` | Chrome/Safari UA | Override the User-Agent header |
| `--rate` | `500ms` | Min delay between requests |
| `--timeout` | `30s` | Per-request timeout |
| `--retries` | `3` | Retry attempts on 429 or 5xx |
| `-q, --quiet` | off | Suppress progress output on stderr |
| `--color` | auto | `auto`, `always`, or `never` |

## Rate limiting and retries

`--rate` sets the minimum gap between consecutive requests (default 500ms).
On a 429 or 5xx response, `weibo` backs off with exponential delay (500ms ×
attempt, capped at 5s) and retries up to `--retries` times before giving up.

## Output auto-detection

The default output format adapts to where it is going: an aligned table when
writing to a terminal, JSONL when piped. That keeps interactive use readable
and scripted use parseable without setting `-o` either way. See
[output formats](/reference/output/) for the full set.
