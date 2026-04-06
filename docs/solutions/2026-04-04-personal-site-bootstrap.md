# Personal Blog & Site — Bootstrap

**Date**: 2026-04-04
**Type**: feature

## Summary

A Go static site generator that reads markdown content from `content/` and outputs a complete static website to `dist/`. It is a single binary with embedded templates and static assets, deployed by copying `dist/` to any static host.

## How the Site Works

### Build vs. Serve

Two modes, both driven by `main.go`:

- `go run . build` — reads all content, renders HTML, writes `dist/`
- `go run . serve` — runs an HTTP file server on `dist/` + a file watcher that rebuilds on any change, with live-reload via a long-poll endpoint injected into served HTML

### Content Flow

```
content/posts/*.md   ─┐
content/projects/*.md ─┼─ LoadPosts/LoadProjects ─► splitFrontmatter ─► goldmark.Convert ─► HTML
content/about.md     ─┘

site.yaml ────────────────────────────────────────────────────────────── SiteConfig
```

Each markdown file has YAML frontmatter (`---` delimited) parsed separately from the body. The body is then rendered by goldmark (CommonMark + GFM + Chroma syntax highlighting). Image paths like `./images/foo.png` are rewritten to `/images/foo.png` via regex.

### Template Architecture

Each page gets its **own independent `*template.Template`**, keyed in a `map[string]*template.Template`. Each template is built by concatenating `base.html` (layout) + `page.html` (content block), then parsing the combined string.

This is critical: do NOT share a single `*template.Template` across all pages. Go's `html/template` keeps all `{{define "content"}}` blocks in one namespace — the last parsed definition wins for all pages.

### Page Data

All pages receive the same `PageData` struct (`internal/builder/builder.go`), which carries:
- `Site` — full `SiteConfig` (title, author, social links, etc.)
- `CurrentPath` — used in `base.html` to highlight the active nav link
- `Year` — for the footer copyright
- Page-specific fields: `Posts`, `Post`, `Projects`, `About`, `AllTags`, `ActiveTag`, `Page`, `TotalPages`

### Output Structure

```
dist/
  index.html                      ← home
  blog/
    index.html                    ← blog list (page 1)
    page/2/index.html             ← paginated (page 2+)
    <slug>/index.html             ← individual posts
    tag/<tag>/index.html          ← tag-filtered list
  projects/index.html
  about/index.html
  feed.xml                        ← RSS 2.0
  search-index.json               ← search index (title, slug, tags, excerpt, content preview)
  sitemap.xml
  robots.txt
  images/                         ← copied from content/images/
  static/js/app.js                ← dark mode + search JS
```

All pages use "pretty URLs" (`/blog/my-post/index.html` serves as `/blog/my-post/`).

### Dark/Light Mode

Tailwind's `darkMode: 'class'` strategy is used. A small inline `<script>` in `<head>` (before body render) reads `localStorage.theme` and sets/removes the `dark` class on `<html>` immediately, preventing flash of wrong theme (FOUC). The toggle button in the header calls `app.js` which flips the class and saves the preference.

Chroma (syntax highlighter) uses the `github` theme by default, with `.dark .chroma` CSS overrides for dark mode colors — no separate stylesheet is generated.

### Client-Side Search

At build time, `search-index.json` is generated with an entry per post: `{title, slug, tags, excerpt, content_preview}`. On the blog page, `app.js` filters the already-rendered `<article class="post-item">` elements by matching the query against `data-title`, `data-tags`, and `data-excerpt` attributes (HTML data attributes on each article, set at build time). No network request at search time — the DOM is the source of truth.

### Embed Strategy

`//go:embed` directives must be in the package that owns the files. Since `templates/` and `static/` are at the repo root, the embed declarations live in `main.go`, and both `embed.FS` values are passed as parameters to `builder.Build()` and `server.Serve()`. Template files are read from the embedded FS at build time; they are not hot-reloadable without recompiling.

## What Changed

- `main.go` — CLI entrypoint, embed declarations, `build`/`serve` dispatch
- `internal/config/config.go` — `SiteConfig` struct + YAML loader with defaults
- `internal/content/content.go` — `Post`, `Project`, `AboutPage` models; `LoadPosts`, `LoadProjects`, `LoadAbout`; manual frontmatter splitting
- `internal/content/markdown.go` — goldmark instance (GFM + Chroma + unsafe HTML); `RenderMarkdown`, render helpers, image path rewriting
- `internal/builder/builder.go` — full build pipeline; `map[string]*template.Template` template parsing; `PageData` struct; all page generation
- `internal/builder/rss.go` — RSS 2.0 feed via `encoding/xml`
- `internal/builder/search.go` — `search-index.json` generation; HTML stripping for content preview
- `internal/builder/sitemap.go` — `sitemap.xml` via `encoding/xml`
- `internal/server/server.go` — HTTP file server; `fsnotify` watcher; long-poll live reload; reload script injection via `</body>` replacement
- `templates/*.html` — 7 templates (base + 6 pages); Tailwind CDN with typography plugin; Inter + JetBrains Mono fonts
- `static/js/app.js` — dark mode toggle; mobile menu; client-side search with 150ms debounce
- `content/` — sample post, project, about page
- `site.yaml` — site configuration file

## Key Decisions

- **`map[string]*template.Template` over shared template**: Go's template namespace collapses all `{{define}}` blocks globally. Each page needs its own independent template instance parsed from `base.html + page.html`.
- **Manual frontmatter splitting over goldmark-frontmatter extension**: Simpler, no coupling between metadata extraction and markdown rendering. Just split on `\n---`, unmarshal YAML, render the rest.
- **Embed in `main.go`, pass as parameters**: `//go:embed` cannot reference paths outside its own package. Rather than restructuring the repo, the FS values are passed as parameters down to builder and server.
- **Tailwind v4 CDN over standalone CLI**: No Node.js build step. Single `<script src="https://cdn.tailwindcss.com?plugins=typography">` tag. Accepted tradeoff: ~100KB JS overhead on each page, acceptable for a personal site. Can migrate to standalone CLI later for production optimisation.
- **Long-poll live reload over WebSocket**: Simpler to implement in a standard `net/http` handler. The server holds the request open until a rebuild completes, then returns 200. The client polls again after reload.
- **Pretty URLs**: All pages output as `dir/index.html` so paths like `/blog/my-post` work natively on any static host without configuration.
- **Slug from filename**: `2026-04-04-my-post.md` → `my-post` (strip date prefix and `.md`). Dates stay in the frontmatter for display.

## Discoveries

- **Go template namespace collision**: When multiple `{{define "content"}}` blocks are parsed into a single `*template.Template`, the last one wins globally. Always parse each page template into its own independent instance.
- **`//go:embed` path restriction**: Paths must be within the same module directory and cannot use `../`. If shared assets live at the repo root, put the embed declarations in the root package (`main.go`).
- **Tailwind CDN config**: `tailwind.config = {...}` must be set in a `<script>` block *before* the CDN script is loaded, or after if using a dedicated config block — but the CDN script must be loaded only once. Loading it twice (plain + `?plugins=typography`) caused double-processing.
- **Chroma dark mode**: Chroma generates HTML with class names (`<span class="kd">`) using its own CSS. The cleanest approach is `.dark .chroma { ... }` overrides in a `<style>` block rather than generating two separate stylesheets.
- **FOUC prevention**: The theme-setting script must be inline in `<head>`, not in a `defer`-loaded external JS file. If it runs after paint, users see a flash from the wrong theme.
- **Live reload and screenshot tools**: The long-poll reload causes the page to refresh as soon as any rebuild completes. This interferes with browser automation tools that scroll and screenshot — the page reloads mid-session. When testing manually this is fine; when automating, serve `dist/` with a plain static server instead.

## Testing / Validation

- `go build ./... && go vet ./...` — clean
- `go run . build` — generates all 12 expected files in `dist/`
- Browser: home, blog list, blog post, projects, about pages all render correctly
- Dark/light toggle: switches instantly, preference persists across page reloads
- Search: filters post list live, shows "No posts found" when no matches
- Tag filter pills: "All" highlighted active, clicking navigates to `/blog/tag/<tag>/`
- RSS: `dist/feed.xml` is valid XML with post entries
- Search index: `dist/search-index.json` contains correct title/slug/tags/excerpt/preview fields
