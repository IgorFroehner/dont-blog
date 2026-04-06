package builder

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/igor/my-go-site/internal/config"
	"github.com/igor/my-go-site/internal/content"
)

const distDir = "dist"

type PageData struct {
	Site             *config.SiteConfig
	PageTitle        string
	PageDescription  string
	CanonicalURL     string
	CurrentPath      string
	IsPost           bool
	Year             int
	Post             *content.Post
	Posts            []content.Post
	RecentPosts      []content.Post
	FeaturedProjects []content.Project
	Projects         []content.Project
	About            *content.AboutPage
	AllTags          []string
	ActiveTag        string
	Page             int
	TotalPages       int
}

var funcMap = template.FuncMap{
	"add":      func(a, b int) int { return a + b },
	"subtract": func(a, b int) int { return a - b },
}

func Build(configPath string, templateFS, staticFS embed.FS) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	templates, err := parseTemplates(templateFS)
	if err != nil {
		return fmt.Errorf("parsing templates: %w", err)
	}

	posts, err := content.LoadPosts("content/posts")
	if err != nil {
		return fmt.Errorf("loading posts: %w", err)
	}

	projects, err := content.LoadProjects("content/projects")
	if err != nil {
		return fmt.Errorf("loading projects: %w", err)
	}

	about, err := content.LoadAbout("content/about.md")
	if err != nil {
		return fmt.Errorf("loading about: %w", err)
	}

	if err := content.RenderAllPosts(posts); err != nil {
		return fmt.Errorf("rendering posts: %w", err)
	}
	if err := content.RenderAllProjects(projects); err != nil {
		return fmt.Errorf("rendering projects: %w", err)
	}
	if err := content.RenderAbout(about); err != nil {
		return fmt.Errorf("rendering about: %w", err)
	}

	if err := os.RemoveAll(distDir); err != nil {
		return fmt.Errorf("cleaning dist: %w", err)
	}

	tagIndex := buildTagIndex(posts)
	allTags := sortedTagNames(tagIndex)
	year := time.Now().Year()

	// Home page
	recentPosts := posts
	if len(recentPosts) > 5 {
		recentPosts = recentPosts[:5]
	}
	var featuredProjects []content.Project
	for _, p := range projects {
		if p.Featured {
			featuredProjects = append(featuredProjects, p)
		}
	}
	if err := renderPage(templates, "home.html", "index.html", PageData{
		Site: cfg, CurrentPath: "/", Year: year,
		RecentPosts: recentPosts, FeaturedProjects: featuredProjects,
	}); err != nil {
		return err
	}

	// Blog pages (paginated)
	if err := renderBlogPages(templates, cfg, posts, allTags, year); err != nil {
		return err
	}

	// Individual posts
	for i := range posts {
		if err := renderPage(templates, "post.html", filepath.Join("blog", posts[i].Slug, "index.html"), PageData{
			Site: cfg, CurrentPath: "/blog", Year: year,
			PageTitle: posts[i].Title, PageDescription: posts[i].Excerpt,
			CanonicalURL: cfg.URL + "/blog/" + posts[i].Slug,
			IsPost: true, Post: &posts[i],
		}); err != nil {
			return err
		}
	}

	// Tag pages
	for tag, tagPosts := range tagIndex {
		if err := renderPage(templates, "tag.html", filepath.Join("blog", "tag", tag, "index.html"), PageData{
			Site: cfg, CurrentPath: "/blog", Year: year,
			PageTitle: "Posts tagged \"" + tag + "\"",
			Posts: tagPosts, ActiveTag: tag, AllTags: allTags,
		}); err != nil {
			return err
		}
	}

	// Projects page
	if err := renderPage(templates, "projects.html", filepath.Join("projects", "index.html"), PageData{
		Site: cfg, CurrentPath: "/projects", Year: year,
		PageTitle: "Projects", Projects: projects,
	}); err != nil {
		return err
	}

	// About page
	if err := renderPage(templates, "about.html", filepath.Join("about", "index.html"), PageData{
		Site: cfg, CurrentPath: "/about", Year: year,
		PageTitle: "About", About: about,
	}); err != nil {
		return err
	}

	// RSS feed
	if err := generateRSS(cfg, posts, filepath.Join(distDir, "feed.xml")); err != nil {
		return fmt.Errorf("generating RSS: %w", err)
	}

	// Search index
	if err := generateSearchIndex(posts, filepath.Join(distDir, "search-index.json")); err != nil {
		return fmt.Errorf("generating search index: %w", err)
	}

	// Copy images
	if err := copyDir("content/images", filepath.Join(distDir, "images")); err != nil {
		return fmt.Errorf("copying images: %w", err)
	}

	// Copy static assets
	if err := copyEmbeddedFS(staticFS, "static", filepath.Join(distDir, "static")); err != nil {
		return fmt.Errorf("copying static: %w", err)
	}

	// Sitemap
	if err := generateSitemap(cfg, posts, projects, tagIndex, filepath.Join(distDir, "sitemap.xml")); err != nil {
		return fmt.Errorf("generating sitemap: %w", err)
	}

	// robots.txt
	if err := writeFile(filepath.Join(distDir, "robots.txt"), fmt.Sprintf("User-agent: *\nAllow: /\nSitemap: %s/sitemap.xml\n", cfg.URL)); err != nil {
		return fmt.Errorf("generating robots.txt: %w", err)
	}

	return nil
}

func parseTemplates(templateFS embed.FS) (map[string]*template.Template, error) {
	baseContent, err := fs.ReadFile(templateFS, "templates/base.html")
	if err != nil {
		return nil, fmt.Errorf("reading base template: %w", err)
	}

	pages := []string{"home.html", "blog.html", "post.html", "projects.html", "about.html", "tag.html"}
	templates := make(map[string]*template.Template, len(pages))

	for _, page := range pages {
		pageContent, err := fs.ReadFile(templateFS, "templates/"+page)
		if err != nil {
			return nil, fmt.Errorf("reading template %s: %w", page, err)
		}

		combined := string(baseContent) + "\n" + string(pageContent)
		t, err := template.New(page).Funcs(funcMap).Parse(combined)
		if err != nil {
			return nil, fmt.Errorf("parsing template %s: %w", page, err)
		}
		templates[page] = t
	}

	return templates, nil
}

func renderPage(templates map[string]*template.Template, templateName, outputPath string, data PageData) error {
	fullPath := filepath.Join(distDir, outputPath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return err
	}

	f, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer f.Close()

	t, ok := templates[templateName]
	if !ok {
		return fmt.Errorf("template %q not found", templateName)
	}
	return t.Execute(f, data)
}

func renderBlogPages(templates map[string]*template.Template, cfg *config.SiteConfig, posts []content.Post, allTags []string, year int) error {
	perPage := cfg.PostsPerPage
	totalPages := (len(posts) + perPage - 1) / perPage
	if totalPages == 0 {
		totalPages = 1
	}

	for page := 1; page <= totalPages; page++ {
		start := (page - 1) * perPage
		end := start + perPage
		if end > len(posts) {
			end = len(posts)
		}

		outputPath := filepath.Join("blog", "index.html")
		if page > 1 {
			outputPath = filepath.Join("blog", "page", fmt.Sprintf("%d", page), "index.html")
		}

		if err := renderPage(templates, "blog.html", outputPath, PageData{
			Site: cfg, CurrentPath: "/blog", Year: year,
			PageTitle: "Blog", Posts: posts[start:end],
			AllTags: allTags, Page: page, TotalPages: totalPages,
		}); err != nil {
			return err
		}
	}

	return nil
}

func buildTagIndex(posts []content.Post) map[string][]content.Post {
	index := make(map[string][]content.Post)
	for _, p := range posts {
		for _, tag := range p.Tags {
			index[tag] = append(index[tag], p)
		}
	}
	return index
}

func sortedTagNames(index map[string][]content.Post) []string {
	tags := make([]string, 0, len(index))
	for tag := range index {
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	return tags
}

func copyDir(src, dst string) error {
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return nil
	}

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)

		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return writeFileBytes(target, data)
	})
}

func copyEmbeddedFS(fsys embed.FS, root, dst string) error {
	return fs.WalkDir(fsys, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, _ := filepath.Rel(root, path)
		target := filepath.Join(dst, rel)

		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}

		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}
		return writeFileBytes(target, data)
	})
}

func writeFile(path, content string) error {
	return writeFileBytes(path, []byte(content))
}

func writeFileBytes(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// stripHTML removes HTML tags from a string.
func stripHTML(s string) string {
	var result strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			result.WriteRune(r)
		}
	}
	return result.String()
}
