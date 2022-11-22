package model

import "encoding/xml"

//PodcastData is
type RssPodcastData struct {
	XMLName    xml.Name   `xml:"rss"`
	Text       string     `xml:",chardata"`
	Itunes     string     `xml:"itunes,attr"`
	Atom       string     `xml:"atom,attr"`
	Media      string     `xml:"media,attr"`
	Psc        string     `xml:"psc,attr"`
	Omny       string     `xml:"omny,attr"`
	Content    string     `xml:"content,attr"`
	Googleplay string     `xml:"googleplay,attr"`
	Acast      string     `xml:"acast,attr"`
	Version    string     `xml:"version,attr"`
	Channel    RssChannel `xml:"channel"`
}
type RssChannel struct {
	Text        string       `xml:",chardata"`
	Language    string       `xml:"language"`
	Link        string       `xml:"link"`
	Title       string       `xml:"title"`
	Description string       `xml:"description"`
	Type        string       `xml:"type"`
	Summary     string       `xml:"summary"`
	Image       RssItemImage `xml:"image"`
	Item        []RssItem    `xml:"item"`
	Author      string       `xml:"author"`
}
type RssItem struct {
	Text        string           `xml:",chardata"`
	Title       string           `xml:"title"`
	Description string           `xml:"description"`
	Encoded     string           `xml:"encoded"`
	Summary     string           `xml:"summary"`
	EpisodeType string           `xml:"episodeType"`
	Author      string           `xml:"author"`
	Image       RssItemImage     `xml:"image"`
	Guid        RssItemGuid      `xml:"guid"`
	ClipId      string           `xml:"clipId"`
	PubDate     string           `xml:"pubDate"`
	Duration    string           `xml:"duration"`
	Enclosure   RssItemEnclosure `xml:"enclosure"`
	Link        string           `xml:"link"`
	Episode     string           `xml:"episode"`
}

type RssItemEnclosure struct {
	Text   string `xml:",chardata"`
	URL    string `xml:"url,attr"`
	Length string `xml:"length,attr"`
	Type   string `xml:"type,attr"`
}
type RssItemImage struct {
	Text string `xml:",chardata"`
	Href string `xml:"href,attr"`
	URL  string `xml:"url"`
}

type RssItemGuid struct {
	Text        string `xml:",chardata"`
	IsPermaLink string `xml:"isPermaLink,attr"`
}
