---
title: "Introduction"
description: "What Weibo's public API looks like and how weibo turns it into pipeable records."
weight: 10
---

[Weibo](https://weibo.com) is one of the largest social platforms in China:
trending topics, posts, comments, user profiles, and timelines, all behind a
web frontend that talks to a JSON API. That API is public in the sense that
your browser uses it every time you load a page, but it is not built to be
called by hand.

`weibo` closes that gap. It is a single binary that treats Weibo the way
`curl` treats a web server: you ask for something by its id or URL, it fetches
exactly that, and it hands you a clean structured record.

## What the API looks like

Almost every endpoint returns the same envelope:

```json
{ "ok": 1, "data": { ... } }
```

A value of `1` means success and `data` carries the payload. A value of `-100`
means the surface requires a logged-in session. `weibo` maps those codes to
clear exit statuses, so you know exactly what went wrong.

## Two hosts, different headers

Weibo's API is split across two hosts:

- **weibo.com** — hot search board and search suggestions. Needs desktop
  Chrome headers and `Referer: https://weibo.com/`.
- **m.weibo.cn** — post detail, comments, user profiles, and timelines. Needs
  mobile Safari headers plus `MWeibo-Pwa: 1` and
  `X-Requested-With: XMLHttpRequest`.

`weibo` handles that routing for you based on which command you run.

## What the six commands read

| Command | Endpoint | Needs cookie? |
|---|---|---|
| `hot` | `weibo.com/ajax/side/hotSearch` | no |
| `status` | `m.weibo.cn/statuses/show` | no |
| `comments` | `m.weibo.cn/comments/hotflow` | no |
| `suggest` | `weibo.com/ajax/side/search` | no |
| `user` | `m.weibo.cn/api/container/getIndex?containerid=100505{uid}` | yes |
| `posts` | `m.weibo.cn/api/container/getIndex?containerid=107603{uid}` | yes |

The four open commands work without any account. The two gated commands return
ok:-100 for anonymous callers — pass your browser `SUB` cookie with
`--cookie` or `WEIBO_COOKIE` to reach them.

## What weibo cleans up for you

**HTML in post text.** Post bodies contain `<a>`, `<span>`, and `<br />`
tags. `weibo` replaces line-break tags with spaces, strips all other HTML
tags, and collapses whitespace into clean readable text.

**Timestamps.** Weibo timestamps arrive as `"Mon Jun 15 09:05:12 +0800 2026"`.
`weibo` normalizes them to `"2006-01-02 15:04:05"` UTC.

**ID forms.** `status`, `comments` accept a bare numeric id or any of the
common URL forms: `m.weibo.cn/detail/ID`, `m.weibo.cn/status/ID`. `user`,
`posts` accept a bare uid or `weibo.com/u/UID`, `m.weibo.cn/u/UID`,
`m.weibo.cn/profile/UID`.

**Retries.** On 429 or 5xx the client backs off (500ms per attempt, capped
at 5s) and retries up to `--retries` times (default 3).

## What weibo is not

`weibo` is a read-only client. It does not log in for you, post, or vote. It
reads the public data and shapes it. That narrow scope is what keeps it a
single small binary with no database, no daemon, and no setup.

Next: [install it](/getting-started/installation/), then take the
[quick start](/getting-started/quick-start/).
