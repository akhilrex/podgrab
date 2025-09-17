package controllers

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/akhilrex/podgrab/db"
	"github.com/akhilrex/podgrab/model"
	"github.com/akhilrex/podgrab/service"
	"github.com/gin-gonic/gin"
)

type SearchGPodderData struct {
	Q            string `binding:"required" form:"q" json:"q" query:"q"`
	SearchSource string `binding:"required" form:"searchSource" json:"searchSource" query:"searchSource"`
}
type SettingModel struct {
	DownloadOnAdd                 bool   `form:"downloadOnAdd" json:"downloadOnAdd" query:"downloadOnAdd"`
	InitialDownloadCount          int    `form:"initialDownloadCount" json:"initialDownloadCount" query:"initialDownloadCount"`
	AutoDownload                  bool   `form:"autoDownload" json:"autoDownload" query:"autoDownload"`
	AppendDateToFileName          bool   `form:"appendDateToFileName" json:"appendDateToFileName" query:"appendDateToFileName"`
	AppendEpisodeNumberToFileName bool   `form:"appendEpisodeNumberToFileName" json:"appendEpisodeNumberToFileName" query:"appendEpisodeNumberToFileName"`
	PassthroughPodcastGuid        bool   `form:"passthroughPodcastGuid" json:"passthroughPodcastGuid" query:"passthroughPodcastGuid"`
	DarkMode                      bool   `form:"darkMode" json:"darkMode" query:"darkMode"`
	DownloadEpisodeImages         bool   `form:"downloadEpisodeImages" json:"downloadEpisodeImages" query:"downloadEpisodeImages"`
	GenerateNFOFile               bool   `form:"generateNFOFile" json:"generateNFOFile" query:"generateNFOFile"`
	DontDownloadDeletedFromDisk   bool   `form:"dontDownloadDeletedFromDisk" json:"dontDownloadDeletedFromDisk" query:"dontDownloadDeletedFromDisk"`
	BaseUrl                       string `form:"baseUrl" json:"baseUrl" query:"baseUrl"`
	MaxDownloadConcurrency        int    `form:"maxDownloadConcurrency" json:"maxDownloadConcurrency" query:"maxDownloadConcurrency"`
	UserAgent                     string `form:"userAgent" json:"userAgent" query:"userAgent"`
}

var searchOptions = map[string]string{
	"itunes":       "iTunes",
	"podcastindex": "PodcastIndex",
}
var searchProvider = map[string]service.SearchService{
	"itunes":       new(service.ItunesService),
	"podcastindex": new(service.PodcastIndexService),
}

func AddPage(c *gin.Context) {
	setting := c.MustGet("setting").(*db.Setting)
	c.HTML(http.StatusOK, "addPodcast.html", gin.H{"title": "Add Podcast", "setting": setting, "searchOptions": searchOptions})
}
func HomePage(c *gin.Context) {
	//var podcasts []db.Podcast
	podcasts := service.GetAllPodcasts("")
	setting := c.MustGet("setting").(*db.Setting)
	c.HTML(http.StatusOK, "index.html", gin.H{"title": "Podgrab", "podcasts": podcasts, "setting": setting})
}
func PodcastPage(c *gin.Context) {
	var searchByIdQuery SearchByIdQuery
	if c.ShouldBindUri(&searchByIdQuery) == nil {

		var podcast db.Podcast

		if err := db.GetPodcastById(searchByIdQuery.Id, &podcast); err == nil {
			var pagination model.Pagination
			if c.ShouldBindQuery(&pagination) == nil {
				var page, count int
				if page = pagination.Page; page == 0 {
					page = 1
				}
				if count = pagination.Count; count == 0 {
					count = 10
				}
				setting := c.MustGet("setting").(*db.Setting)
				totalCount := len(podcast.PodcastItems)
				totalPages := int(math.Ceil(float64(totalCount) / float64(count)))
				nextPage, previousPage := 0, 0
				if page < totalPages {
					nextPage = page + 1
				}
				if page > 1 {
					previousPage = page - 1
				}

				from := (page - 1) * count
				to := page * count
				if to > totalCount {
					to = totalCount
				}
				c.HTML(http.StatusOK, "episodes.html", gin.H{
					"title":          podcast.Title,
					"podcastItems":   podcast.PodcastItems[from:to],
					"setting":        setting,
					"page":           page,
					"count":          count,
					"totalCount":     totalCount,
					"totalPages":     totalPages,
					"nextPage":       nextPage,
					"previousPage":   previousPage,
					"downloadedOnly": false,
					"podcastId":      searchByIdQuery.Id,
				})
			} else {
				c.JSON(http.StatusBadRequest, err)
			}
		} else {
			c.JSON(http.StatusBadRequest, err)
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	}

}

func getItemsToPlay(itemIds []string, podcastId string, tagIds []string) []db.PodcastItem {
	var items []db.PodcastItem
	if len(itemIds) > 0 {
		toAdd, _ := service.GetAllPodcastItemsByIds(itemIds)
		items = *toAdd

	} else if podcastId != "" {
		pod := service.GetPodcastById(podcastId)
		items = pod.PodcastItems
	} else if len(tagIds) != 0 {
		tags := service.GetTagsByIds(tagIds)
		var tagNames []string
		var podIds []string
		for _, tag := range *tags {
			tagNames = append(tagNames, tag.Label)
			for _, pod := range tag.Podcasts {
				podIds = append(podIds, pod.ID)
			}
		}
		items = *service.GetAllPodcastItemsByPodcastIds(podIds)
	}
	return items
}

func PlayerPage(c *gin.Context) {

	itemIds, hasItemIds := c.GetQueryArray("itemIds")
	podcastId, hasPodcastId := c.GetQuery("podcastId")
	tagIds, hasTagIds := c.GetQueryArray("tagIds")
	title := "Podgrab"
	var items []db.PodcastItem
	var totalCount int64
	if hasItemIds {
		toAdd, _ := service.GetAllPodcastItemsByIds(itemIds)
		items = *toAdd
		totalCount = int64(len(items))
	} else if hasPodcastId {
		pod := service.GetPodcastById(podcastId)
		items = pod.PodcastItems
		title = "Playing: " + pod.Title
		totalCount = int64(len(items))
	} else if hasTagIds {
		tags := service.GetTagsByIds(tagIds)
		var tagNames []string
		var podIds []string
		for _, tag := range *tags {
			tagNames = append(tagNames, tag.Label)
			for _, pod := range tag.Podcasts {
				podIds = append(podIds, pod.ID)
			}
		}
		items = *service.GetAllPodcastItemsByPodcastIds(podIds)
		if len(tagNames) == 1 {
			title = fmt.Sprintf("Playing episodes with tag : %s", (tagNames[0]))
		} else {
			title = fmt.Sprintf("Playing episodes with tags : %s", strings.Join(tagNames, ", "))
		}
	} else {
		title = "Playing Latest Episodes"
		if err := db.GetPaginatedPodcastItems(1, 20, nil, nil, time.Time{}, &items, &totalCount); err != nil {
			fmt.Println(err.Error())
		}
	}
	setting := c.MustGet("setting").(*db.Setting)

	c.HTML(http.StatusOK, "player.html", gin.H{
		"title":          title,
		"podcastItems":   items,
		"setting":        setting,
		"count":          len(items),
		"totalCount":     totalCount,
		"downloadedOnly": true,
	})

}
func SettingsPage(c *gin.Context) {

	setting := c.MustGet("setting").(*db.Setting)
	diskStats, _ := db.GetPodcastEpisodeDiskStats()
	c.HTML(http.StatusOK, "settings.html", gin.H{
		"setting":   setting,
		"title":     "Update your preferences",
		"diskStats": diskStats,
	})

}
func BackupsPage(c *gin.Context) {

	files, err := service.GetAllBackupFiles()
	var allFiles []interface{}
	setting := c.MustGet("setting").(*db.Setting)

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
			"setting": setting,
		})
	} else {
		c.JSON(http.StatusBadRequest, err)
	}

}

func getSortOptions() interface{} {
	return []struct {
		Label, Value string
	}{
		{"Release (asc)", "release_asc"},
		{"Release (desc)", "release_desc"},
		{"Duration (asc)", "duration_asc"},
		{"Duration (desc)", "duration_desc"},
	}
}
func AllEpisodesPage(c *gin.Context) {
	var filter model.EpisodesFilter
	c.ShouldBindQuery(&filter)
	filter.VerifyPaginationValues()
	setting := c.MustGet("setting").(*db.Setting)
	podcasts := service.GetAllPodcasts("")
	tags, _ := db.GetAllTags("")
	toReturn := gin.H{
		"title":        "All Episodes",
		"podcastItems": []db.PodcastItem{},
		"setting":      setting,
		"page":         filter.Page,
		"count":        filter.Count,
		"filter":       filter,
		"podcasts":     podcasts,
		"tags":         tags,
		"sortOptions":  getSortOptions(),
	}
	c.HTML(http.StatusOK, "episodes_new.html", toReturn)

}

func AllTagsPage(c *gin.Context) {
	var pagination model.Pagination
	var page, count int
	c.ShouldBindQuery(&pagination)
	if page = pagination.Page; page == 0 {
		page = 1
	}
	if count = pagination.Count; count == 0 {
		count = 10
	}

	var tags []db.Tag
	var totalCount int64
	//fmt.Printf("%+v\n", filter)

	if err := db.GetPaginatedTags(page, count,
		&tags, &totalCount); err == nil {

		setting := c.MustGet("setting").(*db.Setting)
		totalPages := math.Ceil(float64(totalCount) / float64(count))
		nextPage, previousPage := 0, 0
		if float64(page) < totalPages {
			nextPage = page + 1
		}
		if page > 1 {
			previousPage = page - 1
		}
		toReturn := gin.H{
			"title":        "Tags",
			"tags":         tags,
			"setting":      setting,
			"page":         page,
			"count":        count,
			"totalCount":   totalCount,
			"totalPages":   totalPages,
			"nextPage":     nextPage,
			"previousPage": previousPage,
		}
		c.HTML(http.StatusOK, "tags.html", toReturn)
	} else {
		c.JSON(http.StatusBadRequest, err)
	}

}

func Search(c *gin.Context) {
	var searchQuery SearchGPodderData
	if c.ShouldBindQuery(&searchQuery) == nil {
		var searcher service.SearchService
		var isValidSearchProvider bool
		if searcher, isValidSearchProvider = searchProvider[searchQuery.SearchSource]; !isValidSearchProvider {
			searcher = new(service.PodcastIndexService)
		}

		data := searcher.Query(searchQuery.Q)
		allPodcasts := service.GetAllPodcasts("")

		urls := make(map[string]string, len(*allPodcasts))
		for _, pod := range *allPodcasts {
			urls[pod.URL] = pod.ID
		}
		for _, pod := range data {
			_, ok := urls[pod.URL]
			pod.AlreadySaved = ok
		}
		c.JSON(200, data)
	}

}

func GetOmpl(c *gin.Context) {

	usePodgrabLink := c.DefaultQuery("usePodgrabLink", "false") == "true"

	data, err := service.ExportOmpl(usePodgrabLink, getBaseUrl(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request"})
		return
	}
	c.Header("Content-Disposition", "attachment; filename=podgrab-export.opml")
	c.Data(200, "text/xml", data)
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
