package builder

import (
	"encoding/json"
	"strings"

	"github.com/igor/my-go-site/internal/content"
)

type searchEntry struct {
	Title          string   `json:"title"`
	Slug           string   `json:"slug"`
	Tags           []string `json:"tags"`
	Excerpt        string   `json:"excerpt"`
	ContentPreview string   `json:"content_preview"`
}

func generateSearchIndex(posts []content.Post, outputPath string) error {
	entries := make([]searchEntry, 0, len(posts))
	for _, p := range posts {
		preview := contentPreview(string(p.HTMLContent), 200)
		entries = append(entries, searchEntry{
			Title:          p.Title,
			Slug:           p.Slug,
			Tags:           p.Tags,
			Excerpt:        p.Excerpt,
			ContentPreview: preview,
		})
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}

	return writeFile(outputPath, string(data))
}

func contentPreview(html string, maxLen int) string {
	plain := stripHTML(html)
	plain = strings.Join(strings.Fields(plain), " ")
	runes := []rune(plain)
	if len(runes) > maxLen {
		return string(runes[:maxLen]) + "..."
	}
	return plain
}

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
