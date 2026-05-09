# AGENTS.md

## Commands
- `go run . build` builds the static site into `dist/`; each build removes and recreates `dist/`.
- `go run . serve` builds, serves `dist/` on `:1313`, watches `content/`, `templates/`, `static/`, and `site.yaml`, then injects a live-reload script into HTML responses.
- `go run . serve 8080` starts the same dev server on a custom port.
- `go test ./...` is the only focused verification currently available; there are no `*_test.go` files, linter config, formatter config, or task runner.

## Architecture
- `main.go` is the only executable entrypoint; it embeds `templates/*` and `static/*` with `//go:embed` and dispatches only `build` or `serve`.
- `internal/builder.Build` is the build pipeline: load `site.yaml`, parse embedded templates, load content, render Markdown, wipe `dist/`, render pages, generate `feed.xml`, `search-index.json`, `sitemap.xml`, `robots.txt`, copy `content/images`, and copy embedded `static` assets.
- `dist/` is generated and gitignored; do not edit it as source.
- `internal/server` serves files from `dist/`; it is not an app server for dynamic routes.

## Templates And Assets
- Page templates are explicitly listed in `internal/builder.parseTemplates`; adding a new page template requires adding it there and adding a render call.
- `templates/icons.html` and `templates/base.html` are prepended to each parsed page template.
- `templates/about.html` exists but is not parsed or rendered; `/about/` is currently generated as a redirect to `/`.
- Tailwind is loaded from the CDN in `templates/base.html`; there is no local CSS build step.
- Browser behavior lives in `static/js/app.js`; embedded static assets are copied to `dist/static/` during build.

## Content Rules
- Posts are read from `content/posts/*.md`; filenames should be `YYYY-MM-DD-slug.md`, and the date prefix is stripped for the URL slug.
- Post frontmatter uses `title`, `date` in `YYYY-MM-DD`, `tags`, `draft`, and `excerpt`; `draft: true` posts are excluded from all generated output.
- Projects are read from `content/projects/*.md`, sorted by `order`, and `featured: true` controls home-page display.
- `content/about.md` is rendered into home-page content, not a standalone about page.
- Markdown image paths matching `images/`, `./images/`, or `../images/` are rewritten to `/images/`; files under `content/images/` are copied to `dist/images/`.
- Goldmark runs with GFM, footnotes, auto heading IDs, Chroma class-based highlighting, and unsafe HTML enabled.

## Deploy
- `Dockerfile` builds with Go 1.24, runs `./blog build`, then serves the generated `dist/` with nginx.
- `nginx.conf` injects an analytics script using `ANALYTICS_ID` and `ANALYTICS_SCRIPT_URL`; unset values make the injected snippet effectively empty.
- `.github/workflows/build-and-push.yml` only builds/pushes the Docker image on `main` or manual dispatch, then optionally triggers Coolify if both redeploy secrets are set.
