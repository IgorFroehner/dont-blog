# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

A custom static site generator written in Go. It reads Markdown content with YAML frontmatter, renders it through HTML templates, and outputs a complete static site to `dist/`.

## Commands

```bash
go run . build          # Build static site → dist/
go run . serve          # Dev server with live reload on :1313
go run . serve 8080     # Dev server on custom port
```

There are no tests yet. No linter is configured.

## Architecture

**Entry point**: `main.go` — embeds `templates/` and `static/` via `//go:embed`, dispatches to `build` or `serve` subcommands.

**Packages**:
- `internal/config` — loads `site.yaml` into `SiteConfig`
- `internal/content` — parses Markdown files with YAML frontmatter (posts, projects, about page), renders via Goldmark with GFM + syntax highlighting (chroma `github` theme)
- `internal/builder` — orchestrates the full build: loads content, renders templates, generates RSS (`feed.xml`), sitemap, search index (`search-index.json`), robots.txt, and copies static assets
- `internal/server` — dev server using `fsnotify` file watcher with debounced rebuilds and long-poll live reload (injects reload script into HTML responses via `</body>` replacement)

**Build pipeline** (`builder.Build`): config → load content → render markdown → clear `dist/` → render all pages → generate RSS/sitemap/search index → copy images & static assets.

**Content structure**:
- `content/posts/*.md` — blog posts. Filename format: `YYYY-MM-DD-slug.md`. Date prefix is stripped to derive the slug. Draft posts (`draft: true`) are excluded.
- `content/projects/*.md` — project pages, sorted by `order` frontmatter field. `featured: true` shows on home page.
- `content/about.md` — about page content.
- `content/images/` — images referenced from content, copied to `dist/images/`. Image paths in markdown are rewritten to absolute `/images/` paths.

**Templates**: `templates/base.html` is the layout; page templates (`home.html`, `blog.html`, `post.html`, `projects.html`, `about.html`, `tag.html`) are concatenated with it. Template data is the `PageData` struct in `builder.go`.

**Config**: `site.yaml` at project root. Key fields: `title`, `url`, `postsPerPage`, `social` (map), `intro`.
