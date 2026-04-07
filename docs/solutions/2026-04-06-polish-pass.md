# Polish Pass: Syntax Highlighting, Icons, Nav & Footer

**Date**: 2026-04-06
**Type**: feature

## Summary
Fixed broken syntax highlighting (colors were red-heavy and dark mode didn't work), introduced a reusable icon template system, added "About" nav link pointing to home, streamlined the footer, and added social icons (GitHub + X) below the ASCII hero.

## What Changed
- `internal/content/markdown.go`: Switched chroma from inline styles (`WithStyle("github")`) to CSS class output (`chromahtml.WithClasses(true)`). This was the root cause of the broken dark-mode highlighting — inline styles always win over CSS overrides.
- `templates/base.html`: Replaced minimal chroma CSS with full class-based themes (One Light for light mode, Catppuccin Mocha for dark mode). Added "About" link to desktop and mobile nav pointing to `/`. Condensed footer from stacked layout (`py-8`, centered social links above copyright) to a single-row flex layout (`py-4`, copyright left, GitHub icon right). Replaced all inline SVGs with `{{template "icon-*"}}` calls.
- `templates/icons.html` (new): Central icon registry — defines named templates for `icon-github`, `icon-x`, `icon-sun`, `icon-moon`, `icon-menu`, `icon-rss`. Each is a single SVG with `w-5 h-5` sizing.
- `internal/builder/builder.go`: Template parser now reads `icons.html` and prepends it to every template's combined source, making icon templates available everywhere.
- `templates/home.html`: Added GitHub and X social icon links below the ASCII hero, using the new icon templates.
- `site.yaml`: Added `X` social link.
- `content/about.md`: Removed "Get in Touch" section (redundant with social icons on home page).

## Key Decisions
- **CSS classes over inline styles for chroma**: Inline styles from chroma's built-in themes override any CSS (including dark mode rules). Using `WithClasses(true)` moves all color control to the stylesheet, making dark/light mode trivial.
- **Template-based icon system over SVG sprite sheet**: Go's `html/template` already supports named sub-templates. Defining icons as `{{define "icon-x"}}...{{end}}` and calling `{{template "icon-x"}}` is zero-dependency, grep-friendly, and doesn't require any build tooling or client-side JS.
- **Prepend strategy for shared partials**: `icons.html` is prepended to the combined template string before parsing. This avoids changing the template loading architecture — no glob scanning, no partial registry, just string concatenation in the same pattern already used for `base.html + page.html`.

## Discoveries
- Chroma's `github` style is surprisingly red-heavy — it uses red for keywords and strings, making Go code look like an error dump. The style name is misleading if you expect GitHub's actual syntax colors.
- `goldmark-highlighting` already depends on `chroma/v2`, so importing `chroma/v2/formatters/html` for `WithClasses(true)` adds no new dependency — just a new import path.
- Go template `{{template "name"}}` (without a dot argument) works fine for static SVG content that doesn't need data. No need to pass `.` when the template body has no dynamic expressions.

## Testing / Validation
- `go build ./...` compiles cleanly
- `go run . build` generates dist without errors
- Visual verification of syntax highlighting in both light and dark modes
- Icon rendering verified across nav, footer, and home hero section
