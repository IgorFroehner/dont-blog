# Home Page Redesign — ASCII Hero + Merged About

**Date**: 2026-04-06
**Type**: feature

## Summary

Merged the home and about pages into a single landing page. The home page now features an interactive ASCII "IF" hero with a mouse-repel particle effect, the bio content from `content/about.md`, and two navigation cards at the bottom linking to Blog and Projects with live content teasers. The `/about` route redirects to `/` via a static meta-refresh page.

## What Changed

- `internal/builder/builder.go`: Added `About: about` to the home page `PageData` so it receives bio content. Replaced the about page template render with a raw HTML redirect file (`<meta http-equiv="refresh">`). Added `aboutRedirectHTML` const. Removed `about.html` from the template parsing list.
- `templates/home.html`: Full rewrite — ASCII `<pre id="ascii-hero">` hero (hidden below `sm:` breakpoint), intro paragraph from `site.yaml`, rendered bio from `about.md`, and two nav cards (Blog with 2 recent post titles, Projects with featured project names).
- `templates/base.html`: Removed "About" link from desktop and mobile navigation.
- `static/js/app.js`: Added ASCII mouse-repel effect — wraps each character in a `<span data-char>`, tracks mouse position via `requestAnimationFrame`-throttled `mousemove`, replaces nearby characters with random Unicode glyphs, restores them with staggered timeouts on leave. Recalculates positions on resize.

## Key Decisions

- **Static meta-refresh redirect over server redirect**: The site is statically hosted, so `/about/index.html` uses `<meta http-equiv="refresh" content="0;url=/">` with a `<link rel="canonical">`. Works on any static host without server config.
- **Raw HTML for redirect instead of template**: The about redirect doesn't need the full `base.html` layout (nav, footer, etc.), so it's written as a raw HTML const in `builder.go` rather than going through the template system. This avoids parsing an unnecessary template.
- **Reusing existing PageData fields**: `PageData.About`, `.RecentPosts`, and `.FeaturedProjects` already existed on the struct — only needed to populate `About` for the home page. No struct changes required.
- **about.html kept on disk**: The template file still exists but is excluded from the `pages` parsing list. Harmless to keep, avoids breaking the embed directive.

## Discoveries

- **Repel radius tuning**: The initial 80px radius scattered too many characters at the large font size (`text-2xl`/`text-3xl`). Tuned down to 30px after QA. The "right" radius depends on font size and character density — needs visual testing, not just a constant.
- **Character position caching with getBoundingClientRect**: Positions are cached once and recalculated on resize. Since the ASCII hero is at the top of the page (no scroll offset issues) and the character count is small (~60 spans), this is sufficient. For a larger grid or scrollable container, you'd need to recache on scroll too.
- **Span wrapping preserves whitespace**: When replacing `<pre>` text content with spans, newlines must remain as raw text nodes (not wrapped in spans) to preserve the line breaks. Spaces within lines are wrapped in spans like any other character — they participate in the scatter effect.

## Testing / Validation

- `go build ./...` and `go vet ./...` — clean
- `go run . build` — generates all expected files in `dist/`
- Browser: ASCII hero renders centered, mouse effect works (scatter + staggered restore)
- `/about` redirects to `/` immediately
- Nav bar has no "About" link (desktop and mobile)
- Dark mode: ASCII art uses accent color, cards and bio render correctly
- Mobile viewport: ASCII art hidden, nav cards stack vertically
