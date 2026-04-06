# Plan: Personal Blog & Site (Go Static Site Generator)

## Overview

Build a Go static site generator that reads markdown content and outputs a complete static website. The generator is a single embedded binary with Go templates, Tailwind CSS via CDN, Inter font, dark/light mode, and blog features (tags, RSS, search, syntax highlighting).

## Steps

### Step 1: Project Scaffolding & Dependencies

**Files**: `go.mod`, `main.go`, directory structure
**Changes**:
- Initialize Go module: `go mod init github.com/igor/my-go-site` (or preferred module path)
- Create directory structure:
  ```
  main.go                    # CLI entrypoint (build/serve commands)
  internal/
    config/config.go         # Site config parsing
    content/content.go       # Markdown/frontmatter parsing, content models
    builder/builder.go       # Build pipeline orchestration
    server/server.go         # Dev server with file watching
  templates/
    base.html                # Base layout (header, footer, head)
    home.html
    blog.html
    post.html
    projects.html
    about.html
    tag.html                 # Posts filtered by tag
  static/
    js/app.js                # Dark mode toggle + search
  content/
    posts/                   # User's markdown posts
    projects/                # User's markdown projects
    about.md
    images/
  site.yaml                  # Site configuration
  ```
- Add dependencies:
  - `github.com/yuin/goldmark` — markdown parsing
  - `github.com/yuin/goldmark-highlighting/v2` — syntax highlighting via Chroma
  - `go.abhg.dev/goldmark/frontmatter` — YAML frontmatter extraction
  - `github.com/fsnotify/fsnotify` — file watching for dev server
  - `gopkg.in/yaml.v3` — site.yaml config parsing
- Wire up `main.go` with two subcommands: `build` and `serve`

**Notes**: Use `cobra` or just `os.Args` for CLI. Given simplicity (2 commands), plain `os.Args` with a switch is enough — no need for a CLI framework.

### Step 2: Config & Content Models

**Files**: `internal/config/config.go`, `internal/content/content.go`
**Changes**:

`config.go`:
- Define `SiteConfig` struct matching `site.yaml` schema (title, description, author, url, intro, social map, postsPerPage)
- `Load(path string) (*SiteConfig, error)` — reads and parses YAML
- Provide sensible defaults (postsPerPage=10)

`content.go`:
- Define `Post` struct: Title, Slug, Date, Tags, Draft, Excerpt, Content (raw markdown), HTMLContent (rendered), next/prev post pointers
- Define `Project` struct: Title, Description, Tags, URL, Demo, Featured, Order, Content, HTMLContent
- Define `AboutPage` struct: HTMLContent
- `LoadPosts(dir string) ([]Post, error)` — reads all `.md` files from `content/posts/`, parses frontmatter, sorts by date desc, excludes drafts
- `LoadProjects(dir string) ([]Project, error)` — reads all `.md` from `content/projects/`, sorts by order field
- `LoadAbout(path string) (*AboutPage, error)` — reads `content/about.md`
- Slug derived from filename (strip date prefix and `.md`), e.g. `2026-04-04-my-first-post.md` → `my-first-post`

**Notes**: Frontmatter parsing uses goldmark-frontmatter extension. Parse YAML frontmatter into the structs.

### Step 3: Markdown Rendering Pipeline

**Files**: `internal/content/markdown.go`
**Changes**:
- Create a configured goldmark instance with:
  - CommonMark + GFM extensions (tables, strikethrough, task lists)
  - Frontmatter extension
  - Chroma syntax highlighting extension — use `monokai` for dark, `github` for light (render both, toggle with CSS class)
  - Unsafe HTML passthrough (to allow raw HTML in markdown if needed)
- `RenderMarkdown(source []byte) (string, error)` — converts markdown bytes to HTML string
- Handle image paths: rewrite relative image paths (e.g., `../../images/foo.png`) to `/images/foo.png` in the output HTML

**Notes**: For dual-theme syntax highlighting, the simplest approach is to render code blocks once and use CSS to swap Chroma theme colors based on a `dark` class on `<html>`. Generate a single Chroma CSS file with both themes scoped.

### Step 4: Go HTML Templates

**Files**: `templates/base.html`, `templates/home.html`, `templates/blog.html`, `templates/post.html`, `templates/projects.html`, `templates/about.html`, `templates/tag.html`
**Changes**:

`base.html` — master layout:
- `<!DOCTYPE html>`, lang attribute, `<head>` with: charset, viewport, Inter + JetBrains Mono from Google Fonts, Tailwind CDN script, inline Tailwind config (extend with Inter font family), Chroma CSS, `static/js/app.js` (defer), SEO meta tags (title, description, og:tags), RSS `<link>` autodiscovery
- Sticky header: site name (link to `/`), nav links (Blog, Projects, About), dark/light toggle button (sun/moon SVG icons)
- `<main>` block with centered column (`max-w-3xl mx-auto px-4 sm:px-6`)
- Footer: social links (icons), copyright
- `{{block "content" .}}{{end}}` for page-specific content

`home.html`:
- Intro section from `site.yaml` intro text
- "Recent Posts" — loop over latest 5 posts, render as `<ul>` with title link + date
- "Featured Projects" — loop over featured projects, render as cards

`blog.html`:
- Search input at top
- Tag filter pills (all unique tags, link to `/blog/tag/:tag`)
- Post list: title (link), date, tag pills, excerpt
- Pagination nav (Previous / Next links)

`post.html`:
- `<article>` with title `<h1>`, meta line (date + tags), then `{{.HTMLContent}}` wrapped in `prose` class
- Back link to `/blog`

`projects.html`:
- Grid (`grid grid-cols-1 md:grid-cols-2 gap-6`) of project cards
- Each card: title, description, tag pills, repo/demo links

`about.html`:
- `{{.HTMLContent}}` in `prose` class, same as post layout

`tag.html`:
- Same as `blog.html` but filtered to one tag, with tag name in heading

**Notes**: Use Tailwind's `prose` class from Typography plugin (available via CDN) for markdown content. Dark mode: use Tailwind's `dark:` variant with class strategy. The CDN script config sets `darkMode: 'class'`.

### Step 5: Build Pipeline

**Files**: `internal/builder/builder.go`
**Changes**:
- `Builder` struct holds: config, parsed templates, content data
- `New(configPath string) (*Builder, error)` — loads config, parses embedded templates
- `Build() error` — main pipeline:
  1. Clean output dir (`dist/`)
  2. Load all content (posts, projects, about) — can parallelize with goroutines
  3. Render all markdown to HTML
  4. Collect tag index (`map[string][]Post`)
  5. Generate pages:
     - `/index.html` — home
     - `/blog/index.html` — blog list page 1
     - `/blog/page/2/index.html`, etc. — paginated
     - `/blog/:slug/index.html` — each post
     - `/blog/tag/:tag/index.html` — each tag page
     - `/projects/index.html`
     - `/about/index.html`
  6. Generate `/feed.xml` (RSS 2.0)
  7. Generate `/search-index.json` (title, slug, tags, excerpt for each post)
  8. Copy `content/images/` → `dist/images/`
  9. Copy/write embedded static assets (JS) to `dist/static/`
- Template data structs for each page (e.g., `BlogPageData{Posts, Page, TotalPages, Tags}`)
- Use `//go:embed templates/*` and `//go:embed static/*` for embedding

**Notes**: Use `sync.WaitGroup` or `errgroup` for parallel markdown rendering. Output uses "pretty URLs" — `/blog/my-post/index.html` serves as `/blog/my-post/`.

### Step 6: RSS Feed & Search Index

**Files**: `internal/builder/rss.go`, `internal/builder/search.go`
**Changes**:

`rss.go`:
- Generate valid RSS 2.0 XML (`feed.xml`) using `encoding/xml`
- Include: channel title/description/link, each post as an item with title, link, pubDate, description (full HTML content or excerpt)

`search.go`:
- Generate `search-index.json` — JSON array of `{title, slug, tags, excerpt, content_preview}` for each post
- `content_preview`: strip HTML, take first ~200 chars of plain text
- This JSON is loaded by the client-side search JS

**Notes**: RSS date format is RFC 1123Z. Keep search index lightweight — full content search would make the JSON too large for many posts.

### Step 7: Dark/Light Mode & Client-Side JS

**Files**: `static/js/app.js`
**Changes**:

Dark/light mode:
- On page load: check `localStorage.getItem('theme')`, fall back to `prefers-color-scheme` media query
- Set/remove `dark` class on `<html>` element
- Toggle button: flip class + save to `localStorage`
- Add `<script>` in `<head>` (inline, before body) to prevent flash of wrong theme (FOUC): reads preference and sets class immediately

Search:
- Fetch `/search-index.json` on blog page load
- On input, filter entries by matching query against title + tags + content_preview (case-insensitive)
- Re-render the visible post list (hide/show with CSS or DOM manipulation)
- Show "No results found" when empty
- Debounce input (150ms)

**Notes**: The inline FOUC-prevention script must be in `base.html` `<head>`, NOT in `app.js` (which loads deferred). Keep `app.js` small — under 100 lines.

### Step 8: Dev Server with File Watching

**Files**: `internal/server/server.go`
**Changes**:
- HTTP file server serving `dist/` directory
- `fsnotify` watcher on `content/`, `templates/`, `static/`, `site.yaml`
- On file change: trigger full rebuild, log to console
- Inject a small live-reload script into served HTML (WebSocket or simple polling endpoint `/reload` that long-polls until a rebuild happens) — this auto-refreshes the browser
- CLI: `./my-go-site serve` starts watcher + server on `localhost:1313` (or configurable port)

**Notes**: Full rebuild on any change is fine — Go is fast enough for small sites. Live-reload adds nice DX but can be a stretch goal — start with manual refresh and add WebSocket reload if time permits.

### Step 9: Sample Content & Final Polish

**Files**: `content/posts/2026-04-04-hello-world.md`, `content/projects/my-go-site.md`, `content/about.md`, `site.yaml`
**Changes**:
- Create `site.yaml` with placeholder config
- Create a sample "Hello World" post with frontmatter, some markdown content, a code block, and an image reference
- Create a sample project entry
- Create about page content
- Add a sample image in `content/images/`
- Add `<meta>` tags for SEO: title, description, canonical URL, Open Graph tags
- Generate `sitemap.xml` during build (list of all page URLs)
- Ensure `robots.txt` is generated

**Notes**: Sample content serves as documentation — it shows the expected frontmatter format and markdown features.

### Step 10: CSS Refinements & Responsive Behavior

**Files**: `templates/base.html` (Tailwind config), Chroma CSS
**Changes**:
- Fine-tune Tailwind config in CDN script: Inter as default sans font, JetBrains Mono as mono font, custom color palette for light/dark
- Ensure `prose` class styles are correct: `prose-lg` for comfortable reading, dark mode overrides via `dark:prose-invert`
- Mobile hamburger menu: simple JS toggle for nav links on small screens
- Code blocks: `overflow-x-auto`, rounded corners, padding, font-size ~14px
- Tag pills: small rounded badges with accent background
- Project cards: border, rounded, hover shadow transition
- Verify color contrast in both themes (4.5:1 minimum)
- Test on mobile viewport (375px), tablet (768px), desktop (1024px+)

**Notes**: This step is iterative — build and visually verify in the browser. Tailwind CDN makes iteration fast since no rebuild is needed for style changes.

## Testing Strategy

- **Step 1**: `go build .` compiles without errors
- **Step 2**: Write a test that loads sample frontmatter YAML and verifies struct fields parse correctly
- **Step 3**: Write a test that renders a markdown string with code block, image, and table — verify HTML output
- **Step 4**: Manually verify templates render by building the site and opening in browser
- **Step 5**: Run `./my-go-site build`, verify `dist/` has all expected files (`index.html`, `blog/index.html`, `blog/hello-world/index.html`, `projects/index.html`, `about/index.html`, `feed.xml`, `search-index.json`)
- **Step 6**: Validate `feed.xml` with an RSS validator; verify `search-index.json` is valid JSON with expected fields
- **Step 7**: Toggle dark/light mode in browser, verify persistence across refresh. Test search with sample posts.
- **Step 8**: Start dev server, edit a markdown file, verify rebuild triggers and browser shows updated content
- **Step 9-10**: Visual inspection across mobile/tablet/desktop viewports in both themes

## Risks & Open Questions

- **Tailwind CDN in production**: The CDN script adds ~100KB JS overhead. Acceptable for a personal site, but if performance becomes a concern, migrate to Tailwind standalone CLI in a future step.
- **Chroma dual-theme CSS**: Generating CSS that works for both light and dark may require custom Chroma CSS generation scoped to `.dark` class. May need to generate two stylesheets and toggle, or use CSS variables.
- **Live-reload complexity**: WebSocket-based live-reload in the dev server adds complexity. Start with a simple page-refresh approach (polling endpoint) and upgrade if needed.
- **Image path rewriting**: Need to handle various relative path patterns in markdown (`./images/`, `../images/`, `/images/`). Test thoroughly with sample content.
- **Large search index**: If the blog grows to hundreds of posts, the JSON search index could become large. Consider lazy-loading or limiting content preview size.

## Findings

- **Go embed paths**: `//go:embed` directives cannot use `../` relative paths. Moved embed declarations to `main.go` (at repo root) and passed `embed.FS` as parameters to `builder.Build()` and `server.Serve()`.
- **Tailwind CDN double-load**: Initial plan had two CDN script tags. Consolidated to a single `?plugins=typography` URL.
- **Frontmatter parsing**: Used manual `splitFrontmatter()` with `---` delimiters + `gopkg.in/yaml.v3` instead of goldmark-frontmatter extension. Simpler and avoids coupling markdown rendering with metadata extraction.
- **Steps 5 & 6 merged**: RSS, search index, and sitemap generation were implemented alongside the build pipeline since they share output path logic and are tightly coupled.
- **Live reload**: Implemented via long-polling `/___reload` endpoint + response body injection (replaces `</body>` with reload script). No WebSocket needed.
- **Go template namespace collision**: All page templates sharing a single `*template.Template` caused the last-parsed `{{define "content"}}` block (tag.html) to win for all pages, rendering every page as a tag page. Fixed by using `map[string]*template.Template` — one independent template per page, each parsed from `base.html + page.html` combined.
