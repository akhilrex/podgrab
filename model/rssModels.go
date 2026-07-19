package model

import "encoding/xml"

//PodcastData is
type RssPodcastData struct {
	XMLName    xml.Name   `xml:"rss"`
	Text       string     `xml:",chardata"`
	Itunes     string     `xml:"xmlns:itunes,attr"`
	Atom       string     `xml:"xmlns:atom,attr"`
	Media      string     `xml:"xmlns:media,attr"`
	Psc        string     `xml:"xmlns:psc,attr"`
	Omny       string     `xml:"xmlns:omny,attr,omitempty"`
	Content    string     `xml:"xmlns:content,attr"`
	Googleplay string     `xml:"xmlns:googleplay,attr,omitempty"`
	Acast      string     `xml:"xmlns:acast,attr,omitempty"`
	Version    string     `xml:"version,attr"`
	Channel    RssChannel `xml:"channel"`
}
type RssChannel struct {
	Text           string            `xml:",chardata"`
	Language       string            `xml:"language"`
	Link           string            `xml:"link"`
	Title          string            `xml:"title"`
	Description    string            `xml:"description"`
	Image          RssItemImage      `xml:"image"`
	Item           []RssItem         `xml:"item"`
	ItunesAuthor   string            `xml:"itunes:author"`
	ItunesSummary  string            `xml:"itunes:summary"`
	ItunesType     string            `xml:"itunes:type"`
	ItunesExplicit string            `xml:"itunes:explicit"`
	ItunesImage    RssItunesImage    `xml:"itunes:image"`
	ItunesCategory RssItunesCategory `xml:"itunes:category"`
	ItunesOwner    RssItunesOwner    `xml:"itunes:owner"`
}
type RssItem struct {
	Text              string           `xml:",chardata"`
	Title             string           `xml:"title"`
	Description       string           `xml:"description"`
	Encoded           string           `xml:"encoded"`
	Image             RssItemImage     `xml:"image"`
	Guid              RssItemGuid      `xml:"guid"`
	ClipId            string           `xml:"clipId,omitempty"`
	PubDate           string           `xml:"pubDate"`
	Enclosure         RssItemEnclosure `xml:"enclosure"`
	Link              string           `xml:"link"`
	ItunesAuthor      string           `xml:"itunes:author,omitempty"`
	ItunesSummary     string           `xml:"itunes:summary"`
	ItunesEpisodeType string           `xml:"itunes:episodeType"`
	ItunesDuration    string           `xml:"itunes:duration"`
	ItunesImage       RssItunesImage   `xml:"itunes:image"`
	ItunesExplicit    string           `xml:"itunes:explicit"`
	ItunesEpisode     string           `xml:"itunes:episode,omitempty"`
}

// RssItunesImage is the self-closing <itunes:image href="..."/> element -
// distinct from the child-element-based RSS <image> in RssItemImage.
type RssItunesImage struct {
	Href string `xml:"href,attr"`
}

// RssItunesCategory is Apple's required channel-level <itunes:category text="..."/>.
type RssItunesCategory struct {
	Text string `xml:"text,attr"`
}

// RssItunesOwner must always carry both children: parsers such as Castopod's
// php-podcast-parser (PodcastFeed\Tags\Itunes\ItunesOwner) require itunes:email
// and only lazily set itunes:name, so a partially-populated owner blows up
// downstream consumers with "Undefined property ...::$itunes_name".
type RssItunesOwner struct {
	Name  string `xml:"itunes:name"`
	Email string `xml:"itunes:email"`
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
