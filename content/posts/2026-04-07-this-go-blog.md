---
title: "Building This Blog: A Custom Static Site Generator in Go"
date: 2026-04-07
tags: ["go", "projects"]
draft: false
excerpt: "Why I built my own SSG in Go instead of using Hugo, and how it works under the hood."
---

**This post was entirely AI generated**

This post is written in Markdown and rendered by a custom static site generator built in Go.

There are plenty of static site generators out there. Hugo, Jekyll, Eleventy, Astro — all excellent tools. So why build another one?

Because building your own is the whole point. You learn more about Go, you get exactly the features you want, and you end up with something you actually understand top to bottom. This post is both a welcome note and a walkthrough of how this site works.

## What This Generator Supports

The short version:

- **Markdown** with CommonMark and GFM extensions
- **Syntax highlighting** with Chroma
- **Dark/light mode** that respects your system preference
- **RSS feed** for subscribers
- **Client-side search** across all posts

The rest of this post explains how those pieces fit together.

## The Big Picture

The generator reads Markdown files with YAML frontmatter, renders them through HTML templates, and writes a complete static site to a `dist/` directory. Two commands:

```bash
go run . build    # Build the site → dist/
go run . serve    # Dev server with live reload on :1313
```

That's it. No config files beyond a single `site.yaml`. No plugin system. No theme marketplace. Just Go, some templates, and Markdown.

## Project Structure

```
.
├── main.go                  # Entry point, embeds templates/ and static/
├── site.yaml                # Site config (title, author, social links)
├── content/
│   ├── posts/*.md           # Blog posts (YYYY-MM-DD-slug.md)
│   ├── projects/*.md        # Project pages
│   ├── about.md             # About page content
│   └── images/              # Images referenced from content
├── templates/
│   ├── base.html            # Shared layout
│   ├── icons.html           # Reusable SVG icon templates
│   ├── home.html            # Home page
│   ├── blog.html            # Blog listing (paginated)
│   ├── post.html            # Individual post
│   ├── projects.html        # Projects listing
│   └── tag.html             # Posts filtered by tag
├── static/
│   └── js/app.js            # Theme toggle, search, ASCII effect
├── internal/
│   ├── config/              # Loads site.yaml
│   ├── content/             # Parses Markdown + frontmatter
│   ├── builder/             # Orchestrates the full build
│   └── server/              # Dev server with file watching
└── dist/                    # Generated output (gitignored)
```

## Embedding at Compile Time

One thing I really like about Go is `//go:embed`. The templates and static assets are embedded directly into the binary at compile time:

```go
//go:embed templates/*
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS
```

This means `go build` produces a single binary that carries everything it needs. No separate template directory to ship alongside it. The `embed.FS` satisfies the `fs.FS` interface, so it works seamlessly with Go's standard file operations.

## Content Parsing

Content files use YAML frontmatter separated by `---` markers. The parser is hand-written — no external frontmatter library needed:

```go
func splitFrontmatter(data []byte) (frontmatter []byte, body []byte, err error) {
    s := string(data)
    if !strings.HasPrefix(s, "---\n") {
        return nil, data, nil
    }
    end := strings.Index(s[4:], "\n---\n")
    if end == -1 {
        return nil, data, nil
    }
    fm := []byte(s[4 : 4+end])
    rest := []byte(s[4+end+5:])
    return fm, rest, nil
}
```

A blog post looks like this:

```markdown
---
title: "My Post Title"
date: 2026-04-07
tags: ["go", "blog"]
draft: false
excerpt: "A short summary."
---

The actual content goes here.
```

Post filenames follow the pattern `YYYY-MM-DD-slug.md`. The date prefix is stripped to derive the URL slug, so `2026-04-07-this-go-blog.md` becomes `/blog/this-go-blog`. Posts with `draft: true` are excluded from the build entirely.

Posts are sorted by date (newest first) and linked together with `Prev`/`Next` pointers, which the post template uses for navigation between articles.

## Markdown Rendering

Markdown is rendered using [Goldmark](https://github.com/yuin/goldmark) with a few extensions:

- **GFM** (GitHub Flavored Markdown) — tables, strikethrough, autolinks, task lists
- **Footnotes** — because sometimes you need them
- **Syntax highlighting** via [Chroma](https://github.com/alecthomas/chroma) through the goldmark-highlighting extension
- **Auto heading IDs** — each heading gets an `id` attribute for deep linking
- **Unsafe HTML** — so you can embed raw HTML in Markdown when needed

### Syntax Highlighting with CSS Classes

One lesson learned: Chroma can output syntax highlighting in two ways — inline styles or CSS classes. I started with inline styles (the default), but they have a fatal flaw: **inline styles beat CSS in specificity**. If you want dark mode to change your code colors, your CSS overrides simply won't work because the inline `style="color: #..."` always wins.

The fix is to tell Chroma to emit CSS classes instead:

```go
highlighting.NewHighlighting(
    highlighting.WithFormatOptions(
        chromahtml.WithClasses(true),
    ),
),
```

Then you define the colors in your stylesheet, where you have full control over light and dark modes. I use a One Light–inspired palette for light mode and Catppuccin Mocha for dark.

### Image Path Rewriting

Images in Markdown are written with relative paths (`../images/photo.jpg`), but the output needs absolute paths (`/images/photo.jpg`). A regex handles this after rendering:

```go
var imagePathRe = regexp.MustCompile(`(<img\s[^>]*?src=")(?:\.\./)*(?:\./)?images/`)

func RenderMarkdown(source []byte) (template.HTML, error) {
    var buf bytes.Buffer
    if err := md.Convert(source, &buf); err != nil {
        return "", err
    }
    result := imagePathRe.ReplaceAllString(buf.String(), "${1}/images/")
    return template.HTML(result), nil
}
```

## The Template System

Templates use Go's `html/template` package. The architecture is simple: `base.html` defines the layout (head, header, footer), and each page template defines a `{{block "content" .}}` that gets inserted into it.

At parse time, each page template is concatenated with the base and parsed as a single unit:

```go
combined := string(iconsContent) + "\n" + string(baseContent) + "\n" + string(pageContent)
t, err := template.New(page).Funcs(funcMap).Parse(combined)
```

There's also an `icons.html` file that defines reusable SVG icons as named templates:

```html
{{define "icon-github"}}
<svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">...</svg>
{{end}}
```

Then any template can use `{{template "icon-github"}}` instead of pasting the SVG inline. It's prepended to every template during parsing so it's available everywhere.

All templates receive a `PageData` struct with the site config, page metadata, posts, projects, and whatever else the page needs. One struct, every page — Go's zero values mean unused fields are just `nil` or empty without causing errors.

## The Build Pipeline

The `Build` function orchestrates everything in a fixed sequence:

1. Load config from `site.yaml`
2. Parse all templates
3. Load and parse content (posts, projects, about page)
4. Render all Markdown to HTML
5. Clear the `dist/` directory
6. Render each page through its template and write to `dist/`
7. Generate RSS feed (`feed.xml`)
8. Generate search index (`search-index.json`)
9. Generate sitemap (`sitemap.xml`)
10. Generate `robots.txt`
11. Copy images and static assets

### RSS

The RSS feed is built by marshaling Go structs directly into XML using `encoding/xml`. Each post's full HTML content goes into the `<description>` field, so RSS readers get the complete article.

### Client-Side Search

Instead of pulling in a search library, the build generates a JSON index of all posts:

```json
[
  {
    "title": "Post Title",
    "slug": "post-slug",
    "tags": ["go"],
    "excerpt": "Short summary",
    "content_preview": "First 200 characters of text..."
  }
]
```

The blog page has a search input that filters posts client-side by matching the query against titles, tags, and excerpts. No server needed, no external service — just a JSON file and a few lines of JavaScript.

### Sitemap

The sitemap includes all pages: home, blog, projects, about, every individual post, and every tag page. It's a straightforward XML file following the sitemap protocol.

## The Dev Server

The `serve` command starts a development server that watches for file changes and rebuilds automatically:

```go
watcher, _ := fsnotify.NewWatcher()
watchDirs := []string{"content", "templates", "static"}
```

File system events are debounced (100ms) to avoid rebuilding multiple times when your editor saves. After each rebuild, connected browsers are notified to reload.

### Live Reload Without WebSockets

The live reload mechanism is a simple long-poll — no WebSocket library needed. Here's how it works:

1. A tiny script is injected into every HTML response by replacing `</body>` with the script + `</body>`
2. The script polls `GET /___reload`
3. The server holds the request open (up to 30 seconds) until a rebuild happens
4. On rebuild, the server responds with `200`, and the browser reloads
5. If nothing happens for 30 seconds, the server responds with `204` and the client polls again

The injection happens through a response recorder middleware that captures the response body, does the replacement, and then writes the modified version. It only touches HTML responses — CSS, JS, and images pass through untouched.

## Styling

The site uses Tailwind CSS via CDN with the Typography plugin. No build step, no `node_modules`. The Tailwind config is inline in `base.html`:

```javascript
tailwind.config = {
    darkMode: 'class',
    theme: {
        extend: {
            fontFamily: {
                sans: ['Inter', 'system-ui', 'sans-serif'],
                mono: ['JetBrains Mono', 'monospace'],
            },
        },
    },
}
```

Dark mode is class-based (`dark` class on `<html>`), toggled with a button, and persisted in `localStorage`. A small inline script in `<head>` sets the class before the page renders to prevent a flash of wrong colors.

## The ASCII Hero

The home page has an ASCII art "IF" logo that reacts to your mouse cursor. Each character in the `<pre>` block is wrapped in a `<span>` with a `data-char` attribute storing its original value.

On `mousemove`, characters within a radius of the cursor are replaced with random Unicode glyphs (`▓`, `░`, `◈`, `⬡`, etc.) and dimmed. When the cursor moves away, characters restore themselves with staggered `setTimeout` delays, creating a dissolve-and-reform effect.

The effect uses `requestAnimationFrame` to throttle mouse events and caches character positions (recalculated on resize) to keep distance calculations fast. With only ~60 characters in the art, performance is not a concern.

## What It Doesn't Have

Intentionally missing:

- **No bundler** — Tailwind CDN and vanilla JS
- **No npm** — zero Node.js dependencies
- **No database** — everything is files
- **No comments** — if I add them, it'll probably be Giscus or something similar
- **No analytics** — maybe later, probably Plausible or Umami
- **No tests** — not yet, but the build-and-check workflow catches most issues

## Dependencies

The entire project has five direct dependencies:

| Package | Purpose |
|---------|---------|
| `goldmark` | Markdown → HTML |
| `goldmark-highlighting` | Syntax highlighting bridge |
| `chroma` | Syntax highlighting engine |
| `yaml.v3` | YAML frontmatter + config parsing |
| `fsnotify` | File watching for dev server |

That's it. Everything else — templating, HTTP server, XML generation, file I/O — is the Go standard library.

## Why Not Hugo?

Hugo is great. It's fast, mature, and has a huge ecosystem. But:

- I wanted to understand every line of code that builds my site
- I wanted exactly the features I need and nothing else
- I wanted an excuse to write more Go
- The entire codebase is ~600 lines of Go — small enough to hold in my head

If you need a blog today and don't care how it works, use Hugo. If you want to learn how a static site generator works, build one. It's a surprisingly satisfying project.
