package db

import (
	"time"
)

//Podcast is
type Podcast struct {
	Base
	Title string

	Summary string `gorm:"type:text"`

	Author string

	Image string

	URL string

	PodcastItems []PodcastItem
}

//PodcastItem is
type PodcastItem struct {
	Base
	PodcastID string
	Podcast   Podcast
	Title     string
	Summary   string `gorm:"type:text"`

	EpisodeType string

	Duration int

	PubDate time.Time

	FileURL string

	GUID  string
	Image string

	DownloadDate time.Time
	DownloadPath string
}
