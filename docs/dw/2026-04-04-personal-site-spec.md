# Spec: Personal Blog & Site (Go Static Site Generator)

## Summary

A lightweight personal website and blog built as a Go static site generator. Markdown files in the repo are compiled to static HTML at build time. The site features a minimalistic, text-focused design with dark/light mode, Tailwind CSS, and Inter font.

## Requirements

### Static Site Generator (Go)

- CLI tool that reads markdown files and outputs a `dist/` (or `public/`) directory of static HTML, CSS, and assets
- Markdown parsing with full CommonMark support plus extensions: tables, footnotes, task lists, strikethrough
- Image support in markdown — relative paths resolve to a local assets directory (e.g., `content/images/`) and are copied to the output
- Frontmatter parsing (YAML) for metadata on posts and projects
- Hot-reload dev server (`go run . serve` or similar) that watches for file changes and rebuilds
- Single binary build — no external runtime dependencies

### Content Structure

```
content/
  posts/
    2026-04-04-my-first-post.md    # Blog posts
  projects/
    project-name.md                 # Project entries
  about.md                          # About page content
  images/                           # Shared images referenced in markdown
```

**Post frontmatter:**
```yaml
---
title: "Post Title"
date: 2026-04-04
tags: ["go", "web"]
draft: false
excerpt: "Optional short description for the post list"
---
```

**Project frontmatter:**
```yaml
---
title: "Project Name"
description: "Short description"
tags: ["Go", "CLI"]
url: "https://github.com/user/repo"
demo: "https://demo.example.com"   # optional
featured: true                      # shown on home page
order: 1                            # display order
---
```

### Pages

1. **Home (`/`)** — Brief intro text (configurable), latest 3-5 posts, featured projects
2. **Blog (`/blog`)** — Paginated list of all posts (newest first), each showing title, date, tags, excerpt
3. **Post (`/blog/:slug`)** — Full rendered markdown post
4. **Projects (`/projects`)** — All projects as cards/list items
5. **About (`/about`)** — Rendered from `about.md`

### Blog Features

- **Tags**: Each post can have multiple tags. `/blog/tag/:tag` shows filtered posts. Tag list visible on blog page.
- **RSS feed**: Auto-generated `feed.xml` at `/feed.xml` with full post content
- **Client-side search**: Lightweight search (e.g., using a pre-built JSON index at build time) that filters posts by title, tags, and content
- **Syntax highlighting**: Code blocks rendered with syntax highlighting at build time (using a library like Chroma for Go). Support for common languages. Dark/light theme-aware colors.
- **Pagination**: Blog list paginates at ~10 posts per page

### Dark/Light Mode

- Toggle button in the header (sun/moon icon)
- Respects `prefers-color-scheme` on first visit
- Persists user preference in `localStorage`
- Smooth transition between modes (CSS transition on background/text colors)
- All elements (code blocks, images with transparency, cards) must look correct in both modes

### Site Configuration

A single config file (e.g., `site.yaml`) for:
```yaml
title: "Site Name"
description: "Site description for SEO"
author: "Name"
url: "https://example.com"
intro: "Short intro text for the home page"
social:
  github: "https://github.com/user"
  # other links
postsPerPage: 10
```

## Constraints

- **Go only** — no Node.js build step. Tailwind CSS via standalone CLI binary or pre-built CSS
- **Tailwind CSS** — utility-first styling. Use Tailwind's typography plugin (`@tailwindcss/typography`) for markdown prose styling
- **Inter font** — loaded from Google Fonts (with `font-display: swap`) or self-hosted for performance
- **No JavaScript frameworks** — vanilla JS only, and minimal (just for dark mode toggle + search)
- **Static output** — the final site is plain HTML/CSS/JS files deployable to any static host (Netlify, Vercel, GitHub Pages, Cloudflare Pages)
- **Fast builds** — leverage Go's concurrency for parallel markdown processing
- **Accessible** — semantic HTML, proper heading hierarchy, sufficient color contrast in both themes, skip-to-content link
- **Responsive** — mobile-first, works on all screen sizes

## Out of Scope

- Comments system
- Newsletter/email subscription
- CMS or admin interface
- Analytics integration (can be added later via script injection)
- Contact form or any server-side functionality
- Multi-language / i18n support
- Image optimization or resizing at build time

## UI/Design

### Layout & Structure

**Global layout**: Single centered column, max-width 768px, with comfortable horizontal padding (16-24px on mobile, auto-centered on desktop).

**Header** (sticky top, all pages):
- Left: Site name as text link to home
- Right: Navigation links — Blog, Projects, About
- Far right: Dark/light mode toggle icon (sun/moon)
- On mobile: hamburger menu or compact horizontal scroll for nav links

**Footer** (all pages):
- Centered, muted text
- Social icon links (GitHub, etc.)
- Copyright line

**Home page**:
- Intro section: 1-2 sentence greeting/bio, left-aligned
- "Recent Posts" section: Latest 5 posts as a simple list (title + date, no cards)
- "Featured Projects" section: 2-3 featured project cards

**Blog list (`/blog`)**:
- Optional tag filter bar at top
- Posts listed vertically: title (link), date, tags as small pills, 1-line excerpt
- Pagination at bottom (Previous / Next)
- Search input at top of list

**Blog post (`/blog/:slug`)**:
- Title (large, bold)
- Meta line: date + tags
- Markdown content rendered via Tailwind Typography (`prose` class)
- Max-width ~700px for optimal reading line length (~65-75 characters)
- No sidebar, no related posts, no sharing buttons — pure content
- Back link to blog list at bottom

**Projects page (`/projects`)**:
- Grid of project cards (2 columns on desktop, 1 on mobile)
- Each card: title, description, tech tags, links (repo + demo if available)

**About page (`/about`)**:
- Rendered markdown, same prose styling as blog posts

### Interactions

- **Dark/light toggle**: Click toggles immediately. Icon animates between sun/moon. CSS transitions on background and text (~200ms).
- **Search**: Input field filters the visible post list as user types (client-side, instant). Shows "No results" state when empty.
- **Tag filtering**: Clicking a tag pill on the blog page filters to that tag (navigates to `/blog/tag/:tag`). Active tag is visually highlighted.
- **Navigation**: Current page link is visually distinct (underline or bold).
- **Loading states**: N/A (static site, all content is pre-rendered).

### Visual Notes

- **Font**: Inter for all text. Body: 16-18px, line-height 1.6-1.75 for prose.
- **Color palette**: Minimal — near-black text on near-white background (light mode), near-white on near-black (dark mode). Accent color for links and tags (e.g., a single blue or teal).
- **Spacing**: Generous whitespace. Posts feel "airy" — large margins between sections.
- **Code blocks**: Rounded corners, subtle background tint, horizontal scroll on overflow. Monospace font (JetBrains Mono or Fira Code via Google Fonts).
- **Responsive breakpoints**: Mobile-first. Single column below 768px, optional 2-col grid for projects above 768px.
- **Accessibility**: All interactive elements keyboard-navigable. Focus rings visible. Color contrast ratio >= 4.5:1.

## Open Questions

- Should the dev server support live-reload (auto-refresh browser on changes) or just rebuild?
- Preferred deployment target? (Affects if we add a deploy script or GitHub Action)
- Do you want an auto-generated sitemap.xml for SEO?
