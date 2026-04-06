# Spec: Home Page Redesign (Home = About + ASCII Hero)

## Summary
Merge the home and about pages into a single landing page with an interactive ASCII "IF" hero,
bio content, and bottom navigation cards pointing to Blog and Projects. The `/about` route redirects to `/`.

## Requirements
- Home page (`/`) displays: ASCII hero, bio/intro text, and bottom navigation section
- ASCII hero renders large "IF" initials using block characters
- On `mousemove`, characters within a configurable radius of the cursor are replaced with random glyphs (from a symbol set); they restore when the cursor moves away
- The effect is smooth: replacement is instant on entry, restoration is gradual (staggered timeouts)
- `/about` redirects to `/` (HTTP 301 or equivalent in the Go static site builder)
- "About" link is removed from the navigation bar (home and about are now one)
- Bottom of the page has two large navigation cards: one for Blog, one for Projects
- Each card shows: icon or label, short description, and an arrow/link
- The page still renders the intro text from `site.yaml` (`.Site.Intro`)
- Bio content (currently in `content/about.md`) is inlined into the home template

## Constraints
- The site is a Go static site generator — no server-side dynamic behavior
- Tailwind CDN + vanilla JS only (no bundler, no npm)
- The ASCII effect must be implemented in `static/js/app.js`
- Must work in both light and dark mode
- Must not break on mobile (ASCII art can be smaller / hidden on very small screens)
- The `/about` redirect is a static HTML redirect page (`<meta http-equiv="refresh">`)

## Out of Scope
- Animated page transitions
- Any backend changes beyond adding the `/about` redirect page
- Responsive reflow of the ASCII art (it can simply be hidden below `sm:` breakpoint)

## UI/Design

### Layout & Structure

```
┌──────────────────────────────────────────────────────────────┐
│  Header (unchanged): IF logo | Blog | Projects | theme btn   │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│   ██╗ ███████╗          ← ASCII hero, centered              │
│   ██║ ██╔════╝             large, monospace font            │
│   ██║ █████╗                                                 │
│   ██║ ██╔══╝                                                 │
│   ██║ ██║                                                    │
│   ╚═╝ ╚═╝                                                    │
│                                                              │
│   Hi, I'm Igor. I write about software engineering,          │
│   Go, and things I find interesting.           ← intro       │
│                                                              │
│   [About me body text from about.md / expanded bio]          │
│                                                              │
├──────────────────────────────────────────────────────────────┤
│   ┌──────────────────┐   ┌──────────────────┐               │
│   │  Blog            │   │  Projects        │               │
│   │  Writing on Go   │   │  Things I built  │               │
│   │  and engineering │   │                  │               │
│   │  Read posts →    │   │  View projects → │               │
│   └──────────────────┘   └──────────────────┘               │
├──────────────────────────────────────────────────────────────┤
│  Footer (unchanged)                                          │
└──────────────────────────────────────────────────────────────┘
```

### Interactions

**ASCII mouse repel effect:**
- The "IF" art is rendered as a `<pre>` with each character wrapped in a `<span>`
- `mousemove` on the `<pre>` (or `document`) computes distance from cursor to each span
- Spans within radius (~80px) are replaced with a random character from the set:
  `▓ ░ ▒ ╳ ◈ ◉ ⬡ ⬢ ⟡ ✦ ⊕ ⊗` (or similar Unicode symbols)
- Original character is stored as a `data-char` attribute on each span
- On `mouseleave`, or when cursor moves away, each span schedules a restore via `setTimeout`
  with a small random delay (0–300ms) to create a staggered re-crystallization effect
- The original block characters (`█ ╗ ╔ ═ ╝ ╚ ║`) are preserved as the base state

**Navigation cards:**
- Cards have a subtle hover state: border brightens or slight background tint
- Cards link to `/blog` and `/projects` respectively

**Nav change:**
- Remove `<a href="/about">About</a>` from both desktop and mobile nav in `base.html`

### Visual Notes
- ASCII art uses `font-mono` (JetBrains Mono already loaded), large size (`text-2xl` or `text-3xl`)
- ASCII art color: `text-accent` / `text-accent-dark` tinted, or muted gray — to be decided during implementation
- Navigation cards reuse the existing border/card style from the project cards on the current home page
- ASCII art hidden on `xs` screens (below `sm:` breakpoint) to avoid layout issues with narrow viewports
- Scattered characters should use a slightly dimmer or different color to visually distinguish them from stable chars

## Resolved Decisions
- **Bio source**: rendered HTML from `content/about.md` — the home page data struct must include `.About` (same as the about page already does)
- **Nav cards**: show the 1–2 most recent post titles as teasers in the Blog card; Projects card shows the most recent/featured project name + description. This gives the cards live content and looks richer than static copy.
