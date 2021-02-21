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

	LastEpisode *time.Time

	PodcastItems []PodcastItem

	Tags []*Tag `gorm:"many2many:podcast_tags;"`

	DownloadedEpisodesCount  int `gorm:"-"`
	DownloadingEpisodesCount int `gorm:"-"`
	AllEpisodesCount         int `gorm:"-"`
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

	DownloadDate   time.Time
	DownloadPath   string
	DownloadStatus DownloadStatus `gorm:"default:0"`

	IsPlayed bool `gorm:"default:false"`

	BookmarkDate time.Time

	LocalImage string
}

type DownloadStatus int

const (
	NotDownloaded DownloadStatus = iota
	Downloading
	Downloaded
	Deleted
)

type Setting struct {
	Base
	DownloadOnAdd                 bool `gorm:"default:true"`
	InitialDownloadCount          int  `gorm:"default:5"`
	AutoDownload                  bool `gorm:"default:true"`
	AppendDateToFileName          bool `gorm:"default:false"`
	AppendEpisodeNumberToFileName bool `gorm:"default:false"`
	DarkMode                      bool `gorm:"default:false"`
	DownloadEpisodeImages         bool `gorm:"default:false"`
}
type Migration struct {
	Base
	Date time.Time
	Name string
}

type JobLock struct {
	Base
	Date     time.Time
	Name     string
	Duration int
}

type Tag struct {
	Base
	Label       string
	Description string     `gorm:"type:text"`
	Podcasts    []*Podcast `gorm:"many2many:podcast_tags;"`
}

func (lock *JobLock) IsLocked() bool {
	return lock != nil && lock.Date != time.Time{}
}

type PodcastItemStatsModel struct {
	PodcastID      string
	DownloadStatus DownloadStatus
	Count          int
}
