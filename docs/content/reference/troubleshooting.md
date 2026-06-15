---
title: "Troubleshooting"
description: "The things that trip people up, and how to fix each one."
weight: 40
---

Most issues come down to network reality, Weibo's access controls, or a
session cookie that has expired — not a bug.

## Exit 4: surface requires a login

`user` and `posts` return this when no session cookie is set or when the
cookie is expired.

```bash
# Set the cookie
export WEIBO_COOKIE='SUB=xxx; SUBP=yyy'
weibo user 2656274875
```

**How to get a fresh cookie:**

1. Open [weibo.com](https://weibo.com) in Chrome and log in.
2. Press F12, go to **Application → Cookies → weibo.com**.
3. Copy the `SUB` and `SUBP` values.
4. Set `WEIBO_COOKIE='SUB=<value>; SUBP=<value>'` in your shell.

Note: the visitor token flow (`genvisitor2`) returns a `sub`/`subp` pair but
it requires a JS-driven cross-domain activation step that cannot be replicated
in pure HTTP without a browser. There is no workaround — a real logged-in
session is the only option for these surfaces.

## Requests failing or returning 429

Weibo rate-limits like any public site. `weibo` already paces requests and
retries transient failures, but a hard limit still means backing off.

Raise the delay between requests:

```bash
weibo posts 2656274875 --rate 2s
```

Lower any parallelism in scripts that call `weibo` in a loop, and retry later.
A burst of 429 or 5xx responses is the site asking you to slow down, not a
defect.

## Nothing found for something you expected

The public surface is not the whole site. Some data requires JavaScript
rendering, regional access, or a logged-in session, and that part is not
reachable without the right setup. Check that the id is correct, verify that
the content is visible in a private browser window, and confirm you are not
hitting a rate limit.

## The binary is not on your PATH

`go install` puts the binary in `$(go env GOPATH)/bin` (usually `~/go/bin`),
and a release archive leaves it wherever you unpacked it. If your shell cannot
find `weibo`, add that directory to your `PATH`. See
[installation](/getting-started/installation/).

## Seeing what weibo actually sent

Pass `-q=false` to keep progress output even in a script, or run with
`--quiet=false` to confirm which endpoints are being called. There is no
`-v` verbose flag; the progress lines on stderr are the visibility mechanism.
