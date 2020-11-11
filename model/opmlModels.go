package model

import "encoding/xml"

type OpmlModel struct {
	XMLName xml.Name `xml:"opml"`
	Text    string   `xml:",chardata"`
	Version string   `xml:"version,attr"`
	Head    struct {
		Text  string `xml:",chardata"`
		Title string `xml:"title"`
	} `xml:"head"`
	Body struct {
		Text    string `xml:",chardata"`
		Outline []struct {
			Text     string `xml:",chardata"`
			Title    string `xml:"title,attr"`
			AttrText string `xml:"text,attr"`
			Type     string `xml:"type,attr"`
			XmlUrl   string `xml:"xmlUrl,attr"`
			Outline  []struct {
				Text     string `xml:",chardata"`
				AttrText string `xml:"text,attr"`
				Title    string `xml:"title,attr"`
				Type     string `xml:"type,attr"`
				XmlUrl   string `xml:"xmlUrl,attr"`
			} `xml:"outline"`
		} `xml:"outline"`
	} `xml:"body"`
}
