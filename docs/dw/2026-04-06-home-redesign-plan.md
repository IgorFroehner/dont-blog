# Plan: Home Page Redesign (ASCII Hero + Merged About)

## Overview
Merge home and about into a single landing page with an interactive ASCII "IF" hero (mouse-repel effect), bio content from `content/about.md`, and bottom nav cards linking to Blog and Projects with live teasers. The `/about` route becomes a redirect.

## Prior Art
From `docs/solutions/2026-04-04-personal-site-bootstrap.md`:
- **Template architecture**: Each page gets its own `*template.Template` parsed from `base.html + page.html`. Don't share templates across pages.
- **PageData struct**: Already has `About *content.AboutPage`, `RecentPosts`, and `FeaturedProjects` fields — we can populate all three for the home page without changing the struct.
- **Embed strategy**: `templates/` and `static/` are embedded via `main.go` and passed down. No changes needed there.

## Steps

### Step 1: Pass About data to the home page ✅
**Files**: `internal/builder/builder.go`
**Changes**: On line 100-104, add `About: about` to the `PageData` for the home page render call. The `about` variable is already loaded on line 66-69. No struct changes needed — `PageData.About` already exists.

### Step 2: Rewrite home template ✅
**Files**: `templates/home.html`
**Changes**: Replace the entire template content with the new layout:
1. **ASCII hero section**: A centered `<pre id="ascii-hero">` containing the "IF" block art, styled with `font-mono text-2xl sm:text-3xl`, using `text-accent`/`text-accent-dark` color. Wrap it in a `<div class="hidden sm:flex justify-center">` so it's hidden on mobile.
2. **Intro + bio section**: Keep the existing `{{.Site.Intro}}` paragraph. Below it, render `{{.About.HTMLContent}}` inside a `<div class="prose">` block (same pattern as `about.html` line 2-4).
3. **Bottom nav cards section**: A `grid grid-cols-1 sm:grid-cols-2 gap-6` at the bottom with two cards:
   - **Blog card**: Title "Blog", subtitle/description, then list the first 2 items from `{{.RecentPosts}}` as linked titles, plus an "All posts →" link.
   - **Projects card**: Title "Projects", then show the first featured project's name and description from `{{.FeaturedProjects}}`, plus an "All projects →" link.
   - Cards reuse the existing border/rounded/hover style from the current project cards.

### Step 3: Replace about page with redirect ✅
**Files**: `templates/about.html`, `internal/builder/builder.go`
**Changes**:
1. Replace `templates/about.html` content with a minimal redirect page: a `<meta http-equiv="refresh" content="0;url=/">` inside a basic HTML document. Since each template extends `base.html`, instead create a **standalone redirect HTML file** directly in the builder — write a raw HTML string to `dist/about/index.html` instead of using a template. Remove `about.html` from the `pages` slice in `parseTemplates` (line 190).
2. In `builder.go`, replace the about page render call (lines 144-149) with a direct `writeFile(filepath.Join(distDir, "about", "index.html"), redirectHTML)` where `redirectHTML` is a const containing the meta-refresh redirect to `/`.

### Step 4: Remove "About" from navigation ✅
**Files**: `templates/base.html`
**Changes**:
1. Remove the About link from desktop nav (line 107): delete the `<a href="/about" ...>About</a>` element.
2. Remove the About link from mobile menu (line 143): delete the `<a href="/about" ...>About</a>` element.
3. Remove the `CurrentPath "/about"` active-state conditional since it's no longer needed.

### Step 5: Implement ASCII mouse-repel effect ✅
**Files**: `static/js/app.js`
**Changes**: Add a new IIFE at the end of the file that:
1. On `DOMContentLoaded`, find `#ascii-hero`. If not present, return early.
2. Read the text content of the `<pre>`, then replace it with spans: each character becomes `<span data-char="X">X</span>`. Preserve newlines as literal `\n` (not wrapped in spans). Cache all span elements and compute their positions via `getBoundingClientRect`.
3. On `mousemove` over the hero's parent container:
   - Compute cursor position relative to the `<pre>`.
   - For each span within ~80px radius: replace `textContent` with a random glyph from `['▓','░','▒','╳','◈','◉','⬡','⬢','✦','⊕','⊗']`, add a CSS class for dimmer color (`text-gray-500`).
   - For each span outside the radius that is currently displaced: schedule restoration via `setTimeout` with a random delay (0–300ms). Store the timeout ID to avoid stacking.
4. On `mouseleave` from the container: restore all spans to their `data-char` values with staggered timeouts.
5. Use `requestAnimationFrame` to throttle `mousemove` — only process once per frame.
6. Recalculate positions on `resize` (debounced).

### Step 6: Build and verify ✅
**Files**: none (verification step)
**Changes**: Run `go run . build` and inspect `dist/index.html` for the new layout. Open in browser, verify:
- ASCII art renders centered, hidden on mobile viewport
- Mouse hover scatters characters, they reform on leave
- Bio text from about.md appears
- Nav cards show recent posts and featured project
- `/about` redirects to `/`
- Dark mode works throughout
- "About" is gone from nav

## Testing Strategy
- `go build ./...` — compiles cleanly
- `go run . build` — generates dist without errors
- Browser check: home page layout, ASCII effect, dark/light mode
- Navigate to `/about` — should redirect to `/`
- Mobile viewport: ASCII art hidden, cards stack vertically
- Nav bar: no "About" link in desktop or mobile

## Risks & Open Questions
- **Character position caching**: `getBoundingClientRect` positions become stale on scroll/resize. The resize handler mitigates this, but if the hero is in a scrollable container, positions may drift. Likely fine since the hero is at the top of the page.
- **Performance**: The "IF" art is small (~60 chars). Per-frame distance checks for 60 spans is trivially fast. No concern here.
- **about.md content**: Currently sparse (3 short paragraphs). The home page will look thin unless the bio is fleshed out. This is a content issue, not an implementation issue.

## Findings
- No struct changes were needed — `PageData.About` was already on the struct, just wasn't populated for the home page.
- The `about.html` template file still exists on disk but is no longer parsed or used. It could be deleted, but keeping it is harmless since it's excluded from the `pages` slice.
- The redirect approach uses `<meta http-equiv="refresh">` + `<link rel="canonical">` which works for static hosts without server-side redirect support.

## QA Amendments

### Reduce ASCII repel radius
**Feedback**: Effect radius too large — too many characters scatter at once
**Root cause**: `RADIUS = 80` in `static/js/app.js` was too wide for the large font size (`text-2xl`/`text-3xl`)
**Fix**: Changed `RADIUS` from `80` to `50` in `static/js/app.js`
**Impact on plan**: Step 5 spec said "~80px" — actual tuned value is 50px
