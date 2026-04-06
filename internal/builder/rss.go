package builder

import (
	"encoding/xml"
	"time"

	"github.com/igor/my-go-site/internal/config"
	"github.com/igor/my-go-site/internal/content"
)

type rssChannel struct {
	XMLName       xml.Name  `xml:"channel"`
	Title         string    `xml:"title"`
	Link          string    `xml:"link"`
	Description   string    `xml:"description"`
	LastBuildDate string    `xml:"lastBuildDate"`
	Items         []rssItem `xml:"item"`
}

type rssFeed struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Channel rssChannel `xml:"channel"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"`
}

func generateRSS(cfg *config.SiteConfig, posts []content.Post, outputPath string) error {
	var items []rssItem
	for _, p := range posts {
		items = append(items, rssItem{
			Title:       p.Title,
			Link:        cfg.URL + "/blog/" + p.Slug,
			Description: string(p.HTMLContent),
			PubDate:     p.Date.Format(time.RFC1123Z),
			GUID:        cfg.URL + "/blog/" + p.Slug,
		})
	}

	feed := rssFeed{
		Version: "2.0",
		Channel: rssChannel{
			Title:         cfg.Title,
			Link:          cfg.URL,
			Description:   cfg.Description,
			LastBuildDate: time.Now().Format(time.RFC1123Z),
			Items:         items,
		},
	}

	data, err := xml.MarshalIndent(feed, "", "  ")
	if err != nil {
		return err
	}

	return writeFile(outputPath, xml.Header+string(data))
}
