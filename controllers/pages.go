package controllers

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/akhilrex/podgrab/db"
	"github.com/akhilrex/podgrab/service"
	"github.com/gin-gonic/gin"
)

type SearchGPodderData struct {
	Q string `binding:"required" form:"q" json:"q" query:"q"`
}
type SettingModel struct {
	DownloadOnAdd        bool `form:"downloadOnAdd" json:"downloadOnAdd" query:"downloadOnAdd"`
	InitialDownloadCount int  `form:"initialDownloadCount" json:"initialDownloadCount" query:"initialDownloadCount"`
	AutoDownload         bool `form:"autoDownload" json:"autoDownload" query:"autoDownload"`
}

func AddPage(c *gin.Context) {

	c.HTML(http.StatusOK, "addPodcast.html", gin.H{"title": "Add Podcast"})
}
func HomePage(c *gin.Context) {
	//var podcasts []db.Podcast
	podcasts := service.GetAllPodcasts()
	c.HTML(http.StatusOK, "index.html", gin.H{"title": "Podgrab", "podcasts": podcasts})
}
func PodcastPage(c *gin.Context) {
	var searchByIdQuery SearchByIdQuery
	if c.ShouldBindUri(&searchByIdQuery) == nil {

		var podcast db.Podcast

		if err := db.GetPodcastById(searchByIdQuery.Id, &podcast); err == nil {
			setting := c.MustGet("setting").(*db.Setting)
			c.HTML(http.StatusOK, "episodes.html", gin.H{
				"title":          podcast.Title,
				"podcastItems":   podcast.PodcastItems,
				"setting":        setting,
				"page":           1,
				"count":          10,
				"totalCount":     len(podcast.PodcastItems),
				"totalPages":     0,
				"nextPage":       0,
				"previousPage":   0,
				"downloadedOnly": false,
			})
		} else {
			c.JSON(http.StatusBadRequest, err)
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	}

}
func SettingsPage(c *gin.Context) {

	setting := c.MustGet("setting").(*db.Setting)
	c.HTML(http.StatusOK, "settings.html", gin.H{
		"setting": setting,
		"title":   "Update your preferences",
	})

}
func BackupsPage(c *gin.Context) {

	files, err := service.GetAllBackupFiles()
	var allFiles []interface{}

	for _, file := range files {
		arr := strings.Split(file, string(os.PathSeparator))
		name := arr[len(arr)-1]
		subsplit := strings.Split(name, "_")
		dateStr := subsplit[2]
		date, err := time.Parse("2006.01.02", dateStr)
		if err == nil {
			toAdd := map[string]interface{}{
				"date": date,
				"name": name,
				"path": strings.ReplaceAll(file, string(os.PathSeparator), "/"),
			}
			allFiles = append(allFiles, toAdd)
		}
	}

	if err == nil {
		c.HTML(http.StatusOK, "backups.html", gin.H{
			"backups": allFiles,
			"title":   "Backups",
		})
	} else {
		c.JSON(http.StatusBadRequest, err)
	}

}
func AllEpisodesPage(c *gin.Context) {
	var pagination Pagination
	if c.ShouldBindQuery(&pagination) == nil {
		var page, count int
		if page = pagination.Page; page == 0 {
			page = 1
		}
		if count = pagination.Count; count == 0 {
			count = 10
		}
		var podcastItems []db.PodcastItem
		var totalCount int64
		if err := db.GetPaginatedPodcastItems(page, count, pagination.DownloadedOnly, &podcastItems, &totalCount); err == nil {
			setting := c.MustGet("setting").(*db.Setting)
			totalPages := math.Ceil(float64(totalCount) / float64(count))
			nextPage, previousPage := 0, 0
			if float64(page) < totalPages {
				nextPage = page + 1
			}
			if page > 1 {
				previousPage = page - 1
			}
			c.HTML(http.StatusOK, "episodes.html", gin.H{
				"title":          "All Episodes",
				"podcastItems":   podcastItems,
				"setting":        setting,
				"page":           page,
				"count":          count,
				"totalCount":     totalCount,
				"totalPages":     totalPages,
				"nextPage":       nextPage,
				"previousPage":   previousPage,
				"downloadedOnly": pagination.DownloadedOnly,
			})
		} else {
			c.JSON(http.StatusBadRequest, err)
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request"})
	}

}

func Search(c *gin.Context) {
	var searchQuery SearchGPodderData
	if c.ShouldBindQuery(&searchQuery) == nil {
		itunesService := new(service.ItunesService)
		data := itunesService.Query(searchQuery.Q)
		allPodcasts := service.GetAllPodcasts()

		urls := make(map[string]string, len(*allPodcasts))
		for _, pod := range *allPodcasts {
			fmt.Println(pod.URL)
			urls[pod.URL] = pod.ID
		}
		for _, pod := range data {
			_, ok := urls[pod.URL]
			fmt.Println(pod.URL + " " + strconv.FormatBool(ok))
			pod.AlreadySaved = ok
		}
		c.JSON(200, data)
	}

}

func GetOmpl(c *gin.Context) {
	data, err := service.ExportOmpl()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request"})
		return
	}
	c.XML(200, data)
}
func UploadOpml(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	defer file.Close()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request"})
		return
	}

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request"})
		return
	}
	content := string(buf.Bytes())
	err = service.AddOpml(content)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	} else {
		c.JSON(200, gin.H{"success": "File uploaded"})
	}
}

func AddNewPodcast(c *gin.Context) {
	var addPodcastData AddPodcastData
	err := c.ShouldBind(&addPodcastData)

	if err == nil {

		_, err = service.AddPodcast(addPodcastData.Url)
		if err == nil {
			go service.RefreshEpisodes()
			c.Redirect(http.StatusFound, "/")

		} else {

			c.JSON(http.StatusBadRequest, err)

		}
	} else {
		//	fmt.Println(err.Error())
		c.JSON(http.StatusBadRequest, err)
	}

}
