---
title: "Output formats"
description: "The output contract every command shares: formats, fields, and templates."
weight: 30
---

Every command renders through the same formatter. Pick a format with `-o`, or
let `weibo` choose: a table when writing to a terminal, JSONL when piped.

## Formats

```bash
weibo hot -o table     # aligned columns for reading
weibo hot -o list      # each record as a named section
weibo hot -o markdown  # GitHub-flavored pipe table
weibo hot -o jsonl     # one JSON object per line
weibo hot -o json      # a single JSON array
weibo hot -o csv       # spreadsheet-friendly
weibo hot -o tsv       # tab-separated
weibo hot -o url       # just the URL column
weibo hot -o raw       # unformatted bytes
```

| Format | Best for |
|---|---|
| `table` | Reading on a terminal; auto-sized to fit width |
| `list` | Each record as a short named section; streams as records arrive |
| `markdown` | Pasting into a GitHub issue, PR, or README |
| `jsonl` | Piping into `jq` or another tool, one object at a time |
| `json` | Loading a whole result as a JSON array |
| `csv` / `tsv` | Spreadsheets and quick column math |
| `url` | Feeding URLs into other commands |
| `raw` | The underlying bytes from the API response |

## Narrowing columns

Keep only the fields you want:

```bash
weibo hot --fields rank,word,heat
weibo comments 5309997458393240 --fields floor,text,likes
```

`--no-header` drops the header row in `table`, `csv`, and `tsv` output, which
helps when a downstream tool expects bare rows.

## Templating rows

For full control over each line, apply a Go `text/template`. Field names are
the JSON keys with the first letter capitalised:

```bash
weibo hot --template '{{.Rank}}. {{.Word}} ({{.Heat}})'
weibo status 5309997458393240 --template '{{.Username}}: {{.Text}}'
```

## Why auto-detection helps

The default adapts to the destination:

```bash
weibo hot                # a table, because this is a terminal
weibo hot | wc -l        # JSONL, because this is a pipe
weibo hot > out.jsonl    # JSONL, because this is a file
```

You only need `-o` when you want something other than that default.
