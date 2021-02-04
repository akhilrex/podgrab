package controllers

import (
	"encoding/xml"
	"strconv"
	"time"

	"github.com/akhilrex/podgrab/db"
	"github.com/akhilrex/podgrab/service"
	"github.com/gin-gonic/gin"
)

type SubsonicPingResponse struct {
	XMLName xml.Name `xml:"subsonic-response"`
	Text    string   `xml:",chardata"`
	Xmlns   string   `xml:"xmlns,attr"`
	Status  string   `xml:"status,attr"`
	Version string   `xml:"version,attr"`
}

type License struct {
	Text           string `xml:",chardata"`
	Email          string `xml:"email,attr"`
	LicenseExpires string `xml:"licenseExpires,attr"`
	Valid          string `xml:"valid,attr"`
}

type Episode struct {
	Text        string    `xml:",chardata"`
	ID          string    `xml:"id,attr"`
	StreamId    string    `xml:"streamId,attr"`
	ChannelId   string    `xml:"channelId,attr"`
	Title       string    `xml:"title,attr"`
	Description string    `xml:"description,attr"`
	PublishDate time.Time `xml:"publishDate,attr"`
	Status      string    `xml:"status,attr"`
	Parent      string    `xml:"parent,attr"`
	IsDir       string    `xml:"isDir,attr"`
	Year        string    `xml:"year,attr"`
	Genre       string    `xml:"genre,attr"`
	CoverArt    string    `xml:"coverArt,attr"`
	Size        string    `xml:"size,attr"`
	ContentType string    `xml:"contentType,attr"`
	Suffix      string    `xml:"suffix,attr"`
	Duration    string    `xml:"duration,attr"`
	BitRate     string    `xml:"bitRate,attr"`
	Path        string    `xml:"path,attr"`
}

type Channel struct {
	Text             string    `xml:",chardata"`
	ID               string    `xml:"id,attr"`
	URL              string    `xml:"url,attr"`
	Title            string    `xml:"title,attr"`
	Description      string    `xml:"description,attr"`
	CoverArt         string    `xml:"coverArt,attr"`
	OriginalImageUrl string    `xml:"originalImageUrl,attr"`
	Status           string    `xml:"status,attr"`
	ErrorMessage     string    `xml:"errorMessage,attr"`
	Episode          []Episode `xml:"episode"`
}

type Podcasts struct {
	Text    string    `xml:",chardata"`
	Channel []Channel `xml:"channel"`
}

type SubsonicPodcastsResponse struct {
	SubsonicPingResponse
	Podcasts Podcasts `xml:"podcasts"`
}

type SubsonicLicenseResponse struct {
	SubsonicPingResponse
	License License `xml:"license"`
}

func Subsonic_Ping(c *gin.Context) {
	toReturn := &SubsonicPingResponse{
		Version: "1.16.0",
		Status:  "ok",
		Xmlns:   "http://subsonic.org/restapi",
		Text:    "",
	}
	c.XML(200, &toReturn)
}

func Subsonic_License(c *gin.Context) {
	toReturn := &SubsonicLicenseResponse{
		SubsonicPingResponse: SubsonicPingResponse{
			Version: "1.16.0",
			Status:  "ok",
			Xmlns:   "http://subsonic.org/restapi",
			Text:    "",
		},
		License: License{
			Email:          "valid@valid.com",
			Valid:          "true",
			LicenseExpires: time.Now().Add(24 * 30 * 12 * time.Hour).Format(time.RFC3339),
		},
	}
	c.XML(200, &toReturn)
}

func Subsonic_GetPodcasts(c *gin.Context) {
	id := c.DefaultQuery("id", "")
	includeEpisodesString := c.DefaultQuery("includeEpisodes", "true")
	includeEpisodes, _ := strconv.ParseBool(includeEpisodesString)

	var allPodcasts *[]db.Podcast
	if id == "" {
		allPodcasts = service.GetAllPodcasts("")
	} else {
		single := service.GetPodcastById(id)
		allPodcasts = &[]db.Podcast{*single}
	}

	var channels []Channel

	for _, pod := range *allPodcasts {
		channel := Channel{
			ID:               pod.ID,
			URL:              pod.URL,
			Title:            pod.Title,
			Description:      pod.Summary,
			CoverArt:         pod.Image,
			OriginalImageUrl: pod.Image,
			Status:           "completed",
		}

		if includeEpisodes {
			thisPod := service.GetPodcastById(pod.ID)
			var allEpisodes []Episode
			for _, item := range thisPod.PodcastItems {
				episode := Episode{
					ID:          item.ID,
					ChannelId:   pod.ID,
					Title:       item.Title,
					Description: item.Summary,
					PublishDate: item.PubDate,
					Status:      "completed",
					Year:        strconv.Itoa(item.PubDate.Year()),
					Genre:       "Podcast",
					Duration:    strconv.Itoa(item.Duration),
					Path:        item.DownloadPath,
				}

				allEpisodes = append(allEpisodes, episode)
			}
			channel.Episode = allEpisodes
		}
		channels = append(channels, channel)
	}

	toReturn := &SubsonicPodcastsResponse{
		SubsonicPingResponse: SubsonicPingResponse{
			Version: "1.16.0",
			Status:  "ok",
			Xmlns:   "http://subsonic.org/restapi",
			Text:    "",
		},
		Podcasts: Podcasts{
			Channel: channels,
		},
	}
	c.XML(200, &toReturn)
}
