package model

import "encoding/xml"

type OpmlModel struct {
	XMLName xml.Name `xml:"opml"`
	Text    string   `xml:",chardata"`
	Version string   `xml:"version,attr"`
	Head    OpmlHead `xml:"head"`
	Body    OpmlBody `xml:"body"`
}

type OpmlHead struct {
	Text  string `xml:",chardata"`
	Title string `xml:"title"`
}

type OpmlBody struct {
	Text    string        `xml:",chardata"`
	Outline []OpmlOutline `xml:"outline"`
}

type OpmlOutline struct {
	Text     string        `xml:",chardata"`
	AttrText string        `xml:"text,attr"`
	Title    string        `xml:"title,attr"`
	Type     string        `xml:"type,attr"`
	XmlUrl   string        `xml:"xmlUrl,attr"`
	Outline  []OpmlOutline `xml:"outline"`
}
