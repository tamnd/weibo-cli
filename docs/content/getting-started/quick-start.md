---
title: "Quick start"
description: "From an empty terminal to hot topics, post detail, and comments, in a handful of commands."
weight: 30
---

This walks the core commands. Every example here hits live data and finishes
in a second or two. The first four work without any account.

## 1. Hot search

The Weibo trending board, top 10 right now:

```bash
weibo hot -n 10
```

```
rank  word        heat     label  category
1     山姆被约谈  2453012  热              
2     严浩翔贺峻霖  322924  新     综艺
...
```

Pipe the topic URLs:

```bash
weibo hot -o url | head -5
```

## 2. Read a post

Fetch one post by its numeric id:

```bash
weibo status 5309997458393240
```

The text field comes back as plain text (HTML stripped). The `created_at`
field is normalized to UTC:

```
id          5309997458393240
bid         R4c4VzdsQ
text        【时政微视频 | #共产党员习近平#】入党52年。 为人民服务。
created_at  2026-06-15 01:05:12
username    央视新闻
likes       1552
comments    263
reposts     334
url         https://m.weibo.cn/detail/5309997458393240
```

You can also pass a full URL:

```bash
weibo status 'https://m.weibo.cn/detail/5309997458393240'
```

## 3. Pull comments

Hot comments under that post, one per line as JSONL:

```bash
weibo comments 5309997458393240 -n 20 -o jsonl | jq '{floor, text, likes}'
```

## 4. Search suggestions

Autocomplete terms matching a query:

```bash
weibo suggest 山姆 -n 5
```

## 5. User profile and timeline (needs a cookie)

`user` and `posts` read from endpoints that require a Weibo session. Get your
cookie from Chrome DevTools:

1. Open [weibo.com](https://weibo.com) in Chrome and log in.
2. Press F12, go to **Application → Cookies → weibo.com**.
3. Copy the values of `SUB` and `SUBP`.

```bash
export WEIBO_COOKIE='SUB=<value>; SUBP=<value>'
```

Now fetch a profile:

```bash
weibo user 2656274875
```

```
id           2656274875
screen_name  央视新闻
verified     true
verified_for 中央电视台新闻中心官方账号
followers    1.8亿
following    244
posts        688481
```

And their recent posts:

```bash
weibo posts 2656274875 -n 10
weibo posts 2656274875 --page 2   # page 2
```

Without `WEIBO_COOKIE` set, both commands exit 4 with a clear message telling
you what is needed.

## 6. Compose

Output is pipeable by default. Dump comments to a file:

```bash
weibo comments 5309997458393240 -o jsonl > comments.jsonl
wc -l comments.jsonl
```

Or project specific fields:

```bash
weibo hot --fields rank,word,heat -o csv
```

## Where to next

- The [CLI reference](/reference/cli/) lists every command, flag, and default.
- [Configuration](/reference/configuration/) explains the cookie setup and
  environment variables.
- [Troubleshooting](/reference/troubleshooting/) covers the most common
  problems, including expired cookies and rate limits.
