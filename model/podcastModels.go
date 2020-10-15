package model

import "encoding/xml"

//PodcastData is
type PodcastData struct {
	XMLName    xml.Name `xml:"rss"`
	Text       string   `xml:",chardata"`
	Itunes     string   `xml:"itunes,attr"`
	Atom       string   `xml:"atom,attr"`
	Media      string   `xml:"media,attr"`
	Psc        string   `xml:"psc,attr"`
	Omny       string   `xml:"omny,attr"`
	Content    string   `xml:"content,attr"`
	Googleplay string   `xml:"googleplay,attr"`
	Acast      string   `xml:"acast,attr"`
	Version    string   `xml:"version,attr"`
	Channel    struct {
		Text     string `xml:",chardata"`
		Language string `xml:"language"`
		Link     []struct {
			Text string `xml:",chardata"`
			Rel  string `xml:"rel,attr"`
			Type string `xml:"type,attr"`
			Href string `xml:"href,attr"`
		} `xml:"link"`
		Title       string `xml:"title"`
		Description string `xml:"description"`
		Type        string `xml:"type"`
		Summary     string `xml:"summary"`
		Owner       struct {
			Text  string `xml:",chardata"`
			Name  string `xml:"name"`
			Email string `xml:"email"`
		} `xml:"owner"`
		Author    string `xml:"author"`
		Copyright string `xml:"copyright"`
		Explicit  string `xml:"explicit"`
		Category  struct {
			Text     string `xml:",chardata"`
			AttrText string `xml:"text,attr"`
			Category struct {
				Text     string `xml:",chardata"`
				AttrText string `xml:"text,attr"`
			} `xml:"category"`
		} `xml:"category"`
		Image struct {
			Text  string `xml:",chardata"`
			Href  string `xml:"href,attr"`
			URL   string `xml:"url"`
			Title string `xml:"title"`
			Link  string `xml:"link"`
		} `xml:"image"`
		Item []struct {
			Text        string `xml:",chardata"`
			Title       string `xml:"title"`
			Description string `xml:"description"`
			Encoded     string `xml:"encoded"`
			Summary     string `xml:"summary"`
			EpisodeType string `xml:"episodeType"`
			Author      string `xml:"author"`
			Image       struct {
				Text string `xml:",chardata"`
				Href string `xml:"href,attr"`
			} `xml:"image"`
			Content []struct {
				Text   string `xml:",chardata"`
				URL    string `xml:"url,attr"`
				Type   string `xml:"type,attr"`
				Player struct {
					Text string `xml:",chardata"`
					URL  string `xml:"url,attr"`
				} `xml:"player"`
			} `xml:"content"`
			Guid struct {
				Text        string `xml:",chardata"`
				IsPermaLink string `xml:"isPermaLink,attr"`
			} `xml:"guid"`
			ClipId    string `xml:"clipId"`
			PubDate   string `xml:"pubDate"`
			Duration  string `xml:"duration"`
			Enclosure struct {
				Text   string `xml:",chardata"`
				URL    string `xml:"url,attr"`
				Length string `xml:"length,attr"`
				Type   string `xml:"type,attr"`
			} `xml:"enclosure"`
			Link       string `xml:"link"`
			StitcherId string `xml:"stitcherId"`
			Episode    string `xml:"episode"`
		} `xml:"item"`
	} `xml:"channel"`
}
