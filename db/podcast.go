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

	DownloadDate   time.Time
	DownloadPath   string
	DownloadStatus DownloadStatus `gorm:"default:0"`

	IsPlayed bool `gorm:"default:false"`
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
	DownloadOnAdd        bool `gorm:"default:true"`
	InitialDownloadCount int  `gorm:"default:5"`
	AutoDownload         bool `gorm:"default:true"`
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

func (lock *JobLock) IsLocked() bool {
	return lock != nil && lock.Date != time.Time{}
}
