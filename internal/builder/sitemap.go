package builder

import (
	"encoding/xml"
	"fmt"

	"github.com/igor/my-go-site/internal/config"
	"github.com/igor/my-go-site/internal/content"
)

type sitemapURL struct {
	Loc string `xml:"loc"`
}

type sitemapURLSet struct {
	XMLName xml.Name     `xml:"urlset"`
	XMLNS   string       `xml:"xmlns,attr"`
	URLs    []sitemapURL `xml:"url"`
}

func generateSitemap(cfg *config.SiteConfig, posts []content.Post, projects []content.Project, tagIndex map[string][]content.Post, outputPath string) error {
	var urls []sitemapURL

	urls = append(urls, sitemapURL{Loc: cfg.URL})
	urls = append(urls, sitemapURL{Loc: cfg.URL + "/blog"})
	urls = append(urls, sitemapURL{Loc: cfg.URL + "/projects"})
	urls = append(urls, sitemapURL{Loc: cfg.URL + "/about"})

	for _, p := range posts {
		urls = append(urls, sitemapURL{Loc: fmt.Sprintf("%s/blog/%s", cfg.URL, p.Slug)})
	}

	for tag := range tagIndex {
		urls = append(urls, sitemapURL{Loc: fmt.Sprintf("%s/blog/tag/%s", cfg.URL, tag)})
	}

	sitemap := sitemapURLSet{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  urls,
	}

	data, err := xml.MarshalIndent(sitemap, "", "  ")
	if err != nil {
		return err
	}

	return writeFile(outputPath, xml.Header+string(data))
}
