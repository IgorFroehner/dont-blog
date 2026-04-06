package content

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Post struct {
	Title       string
	Slug        string
	Date        time.Time
	Tags        []string
	Draft       bool
	Excerpt     string
	RawContent  []byte
	HTMLContent template.HTML
	Prev        *Post
	Next        *Post
}

type Project struct {
	Title       string
	Description string
	Tags        []string
	URL         string
	Demo        string
	Featured    bool
	Order       int
	RawContent  []byte
	HTMLContent template.HTML
}

type AboutPage struct {
	RawContent  []byte
	HTMLContent template.HTML
}

type postFrontmatter struct {
	Title   string   `yaml:"title"`
	Date    string   `yaml:"date"`
	Tags    []string `yaml:"tags"`
	Draft   bool     `yaml:"draft"`
	Excerpt string   `yaml:"excerpt"`
}

type projectFrontmatter struct {
	Title       string   `yaml:"title"`
	Description string   `yaml:"description"`
	Tags        []string `yaml:"tags"`
	URL         string   `yaml:"url"`
	Demo        string   `yaml:"demo"`
	Featured    bool     `yaml:"featured"`
	Order       int      `yaml:"order"`
}

var datePrefixRe = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}-`)

func LoadPosts(dir string) ([]Post, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var posts []Post
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}

		fm, body, err := splitFrontmatter(data)
		if err != nil {
			return nil, err
		}

		var meta postFrontmatter
		if err := yaml.Unmarshal(fm, &meta); err != nil {
			return nil, err
		}

		if meta.Draft {
			continue
		}

		date, err := time.Parse("2006-01-02", meta.Date)
		if err != nil {
			return nil, fmt.Errorf("post %s: invalid date %q: %w", e.Name(), meta.Date, err)
		}
		slug := slugFromFilename(e.Name())

		posts = append(posts, Post{
			Title:      meta.Title,
			Slug:       slug,
			Date:       date,
			Tags:       meta.Tags,
			Draft:      meta.Draft,
			Excerpt:    meta.Excerpt,
			RawContent: body,
		})
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date.After(posts[j].Date)
	})

	for i := range posts {
		if i > 0 {
			posts[i].Next = &posts[i-1]
		}
		if i < len(posts)-1 {
			posts[i].Prev = &posts[i+1]
		}
	}

	return posts, nil
}

func LoadProjects(dir string) ([]Project, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var projects []Project
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}

		fm, body, err := splitFrontmatter(data)
		if err != nil {
			return nil, err
		}

		var meta projectFrontmatter
		if err := yaml.Unmarshal(fm, &meta); err != nil {
			return nil, err
		}

		projects = append(projects, Project{
			Title:       meta.Title,
			Description: meta.Description,
			Tags:        meta.Tags,
			URL:         meta.URL,
			Demo:        meta.Demo,
			Featured:    meta.Featured,
			Order:       meta.Order,
			RawContent:  body,
		})
	}

	sort.Slice(projects, func(i, j int) bool {
		return projects[i].Order < projects[j].Order
	})

	return projects, nil
}

func LoadAbout(path string) (*AboutPage, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	_, body, err := splitFrontmatter(data)
	if err != nil {
		return nil, err
	}

	return &AboutPage{RawContent: body}, nil
}

func slugFromFilename(name string) string {
	name = strings.TrimSuffix(name, ".md")
	return datePrefixRe.ReplaceAllString(name, "")
}

func splitFrontmatter(data []byte) (frontmatter []byte, body []byte, err error) {
	s := string(data)
	if !strings.HasPrefix(s, "---\n") {
		return nil, data, nil
	}

	end := strings.Index(s[4:], "\n---\n")
	if end == -1 {
		return nil, data, nil
	}

	fm := []byte(s[4 : 4+end])
	rest := []byte(s[4+end+5:])
	return fm, rest, nil
}
