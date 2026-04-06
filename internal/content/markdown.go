package content

import (
	"bytes"
	"html/template"
	"regexp"

	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

var md = newMarkdown()

func newMarkdown() goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Footnote,
			highlighting.NewHighlighting(
				highlighting.WithStyle("github"),
				highlighting.WithFormatOptions(),
			),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)
}

var imagePathRe = regexp.MustCompile(`(<img\s[^>]*?src=")(?:\.\./)*(?:\./)?images/`)

func RenderMarkdown(source []byte) (template.HTML, error) {
	var buf bytes.Buffer
	if err := md.Convert(source, &buf); err != nil {
		return "", err
	}

	result := imagePathRe.ReplaceAllString(buf.String(), "${1}/images/")
	return template.HTML(result), nil
}

func RenderAllPosts(posts []Post) error {
	for i := range posts {
		rendered, err := RenderMarkdown(posts[i].RawContent)
		if err != nil {
			return err
		}
		posts[i].HTMLContent = rendered
	}
	return nil
}

func RenderAllProjects(projects []Project) error {
	for i := range projects {
		rendered, err := RenderMarkdown(projects[i].RawContent)
		if err != nil {
			return err
		}
		projects[i].HTMLContent = rendered
	}
	return nil
}

func RenderAbout(about *AboutPage) error {
	rendered, err := RenderMarkdown(about.RawContent)
	if err != nil {
		return err
	}
	about.HTMLContent = rendered
	return nil
}
