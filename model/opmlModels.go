package model

import (
	"encoding/xml"
	"time"
)

type OpmlModel struct {
	XMLName xml.Name `xml:"opml"`
	Text    string   `xml:",chardata"`
	Version string   `xml:"version,attr"`
	Head    OpmlHead `xml:"head"`
	Body    OpmlBody `xml:"body"`
}
type OpmlExportModel struct {
	XMLName xml.Name       `xml:"opml"`
	Text    string         `xml:",chardata"`
	Version string         `xml:"version,attr"`
	Head    OpmlExportHead `xml:"head"`
	Body    OpmlBody       `xml:"body"`
}

type OpmlHead struct {
	Text  string `xml:",chardata"`
	Title string `xml:"title"`
	//DateCreated time.Time `xml:"dateCreated"`
}
type OpmlExportHead struct {
	Text        string    `xml:",chardata"`
	Title       string    `xml:"title"`
	DateCreated time.Time `xml:"dateCreated"`
}

type OpmlBody struct {
	Text    string        `xml:",chardata"`
	Outline []OpmlOutline `xml:"outline"`
}

type OpmlOutline struct {
	Title    string        `xml:"title,attr"`
	XmlUrl   string        `xml:"xmlUrl,attr"`
	Text     string        `xml:",chardata"`
	AttrText string        `xml:"text,attr"`
	Type     string        `xml:"type,attr"`
	Outline  []OpmlOutline `xml:"outline"`
}
