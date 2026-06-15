---
title: "weibo"
description: "A command line for Weibo (微博). Read hot search topics, posts, comments, user profiles, and timelines. One pure-Go binary, no API key."
heroTitle: "Weibo from the command line"
heroLead: "Read hot search topics, posts, comments, and user profiles. One pure-Go binary, no API key, output that pipes into the rest of your tools."
heroPrimaryURL: "/getting-started/quick-start/"
heroPrimaryText: "Get started"
---

A command line for [weibo.com](https://weibo.com).

```bash
weibo hot                          # trending topics right now
weibo status 5309997458393240      # one post, clean text
weibo comments 5309997458393240    # top comments
weibo suggest 山姆                  # autocomplete
weibo user 2656274875 --cookie … # profile (needs session cookie)
weibo posts 2656274875 --cookie … # timeline (needs session cookie)
```

## Where to go next

- New here? Read the [introduction](/getting-started/introduction/), then
  the [quick start](/getting-started/quick-start/).
- Installing? See [installation](/getting-started/installation/).
- Need every flag? The [CLI reference](/reference/cli/) is the full surface.
