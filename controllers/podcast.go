package controllers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/akhilrex/podgrab/model"
	"github.com/akhilrex/podgrab/service"
	"github.com/gin-contrib/location"

	"github.com/akhilrex/podgrab/db"
	"github.com/gin-gonic/gin"
)

const (
	DateAdded   = "dateadded"
	Name        = "name"
	LastEpisode = "lastepisode"
)

const (
	Asc  = "asc"
	Desc = "desc"
)

type SearchQuery struct {
	Q    string `binding:"required" form:"q"`
	Type string `form:"type"`
}

type PodcastListQuery struct {
	Sort  string `uri:"sort" query:"sort" json:"sort" form:"sort" default:"created_at"`
	Order string `uri:"order" query:"order" json:"order" form:"order" default:"asc"`
}

type SearchByIdQuery struct {
	Id string `binding:"required" uri:"id" json:"id" form:"id"`
}

type AddRemoveTagQuery struct {
	Id    string `binding:"required" uri:"id" json:"id" form:"id"`
	TagId string `binding:"required" uri:"tagId" json:"tagId" form:"tagId"`
}

type PatchPodcastItem struct {
	IsPlayed bool   `json:"isPlayed" form:"isPlayed" query:"isPlayed"`
	Title    string `form:"title" json:"title" query:"title"`
}

type AddPodcastData struct {
	Url string `binding:"required" form:"url" json:"url"`
}
type AddTagData struct {
	Label       string `binding:"required" form:"label" json:"label"`
	Description string `form:"description" json:"description"`
}

func GetAllPodcasts(c *gin.Context) {
	var podcastListQuery PodcastListQuery

	if c.ShouldBindQuery(&podcastListQuery) == nil {
		var order = strings.ToLower(podcastListQuery.Order)
		var sorting = "created_at"
		switch sort := strings.ToLower(podcastListQuery.Sort); sort {
		case DateAdded:
			sorting = "created_at"
		case Name:
			sorting = "title"
		case LastEpisode:
			sorting = "last_episode"
		}
		if order == Desc {
			sorting = fmt.Sprintf("%s desc", sorting)
		}

		c.JSON(200, service.GetAllPodcasts(sorting))
	}
}

func GetPodcastById(c *gin.Context) {
	var searchByIdQuery SearchByIdQuery

	if c.ShouldBindUri(&searchByIdQuery) == nil {

		var podcast db.Podcast

		err := db.GetPodcastById(searchByIdQuery.Id, &podcast)
		fmt.Println(err)
		c.JSON(200, podcast)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	}
}

func PausePodcastById(c *gin.Context) {
	var searchByIdQuery SearchByIdQuery
	if c.ShouldBindUri(&searchByIdQuery) == nil {

		err := service.TogglePodcastPause(searchByIdQuery.Id, true)
		if err != nil {
			c.JSON(http.StatusBadRequest, err)
			return
		}
		c.JSON(200, gin.H{})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	}
}
func UnpausePodcastById(c *gin.Context) {
	var searchByIdQuery SearchByIdQuery
	if c.ShouldBindUri(&searchByIdQuery) == nil {
		err := service.TogglePodcastPause(searchByIdQuery.Id, false)
		if err != nil {
			c.JSON(http.StatusBadRequest, err)
			return
		}
		c.JSON(200, gin.H{})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	}
}

func DeletePodcastById(c *gin.Context) {
	var searchByIdQuery SearchByIdQuery

	if c.ShouldBindUri(&searchByIdQuery) == nil {

		service.DeletePodcast(searchByIdQuery.Id, true)
		c.JSON(http.StatusNoContent, gin.H{})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	}
}

func DeleteOnlyPodcastById(c *gin.Context) {
	var searchByIdQuery SearchByIdQuery

	if c.ShouldBindUri(&searchByIdQuery) == nil {

		service.DeletePodcast(searchByIdQuery.Id, false)
		c.JSON(http.StatusNoContent, gin.H{})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	}
}

func DeletePodcastEpisodesById(c *gin.Context) {
	var searchByIdQuery SearchByIdQuery

	if c.ShouldBindUri(&searchByIdQuery) == nil {

		service.DeletePodcastEpisodes(searchByIdQuery.Id)
		c.JSON(http.StatusNoContent, gin.H{})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	}
}
func DeletePodcasDeleteOnlyPodcasttEpisodesById(c *gin.Context) {
	var searchByIdQuery SearchByIdQuery

	if c.ShouldBindUri(&searchByIdQuery) == nil {

		service.DeletePodcastEpisodes(searchByIdQuery.Id)
		c.JSON(http.StatusNoContent, gin.H{})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	}
}

func GetPodcastItemsByPodcastId(c *gin.Context) {
	var searchByIdQuery SearchByIdQuery

	if c.ShouldBindUri(&searchByIdQuery) == nil {

		var podcastItems []db.PodcastItem

		err := db.GetAllPodcastItemsByPodcastId(searchByIdQuery.Id, &podcastItems)
		fmt.Println(err)
		c.JSON(200, podcastItems)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	}
}

func DownloadAllEpisodesByPodcastId(c *gin.Context) {
	var searchByIdQuery SearchByIdQuery

	if c.ShouldBindUri(&searchByIdQuery) == nil {

		err := service.SetAllEpisodesToDownload(searchByIdQuery.Id)
		fmt.Println(err)
		go service.RefreshEpisodes()
		c.JSON(200, gin.H{})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	}
}

func GetAllPodcastItems(c *gin.Context) {
	var filter model.EpisodesFilter
	err := c.ShouldBindQuery(&filter)
	if err != nil {
		fmt.Println(err.Error())
	}
	filter.VerifyPaginationValues()
	if podcastItems, totalCount, err := db.GetPaginatedPodcastItemsNew(filter); err == nil {
		filter.SetCounts(totalCount)
		toReturn := gin.H{
			"podcastItems": podcastItems,
			"filter":       &filter,
		}
		c.JSON(http.StatusOK, toReturn)
	} else {
		c.JSON(http.StatusBadRequest, err)
	}

}

func GetPodcastItemById(c *gin.Context) {
	var searchByIdQuery SearchByIdQuery

	if c.ShouldBindUri(&searchByIdQuery) == nil {

		var podcast db.PodcastItem

		err := db.GetPodcastItemById(searchByIdQuery.Id, &podcast)
		fmt.Println(err)
		c.JSON(200, podcast)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	}
}

func GetPodcastItemImageById(c *gin.Context) {
	var searchByIdQuery SearchByIdQuery

	if c.ShouldBindUri(&searchByIdQuery) == nil {

		var podcast db.PodcastItem

		err := db.GetPodcastItemById(searchByIdQuery.Id, &podcast)
		if err == nil {
			if _, err = os.Stat(podcast.LocalImage); os.IsNotExist(err) {
				c.Redirect(302, podcast.Image)
			} else {
				c.File(podcast.LocalImage)
			}
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	}
}

func GetPodcastImageById(c *gin.Context) {
	var searchByIdQuery SearchByIdQuery

	if c.ShouldBindUri(&searchByIdQuery) == nil {

		var podcast db.Podcast

		err := db.GetPodcastById(searchByIdQuery.Id, &podcast)
		if err == nil {
			localPath := service.GetPodcastLocalImagePath(podcast.Image, podcast.Title)
			if _, err = os.Stat(localPath); os.IsNotExist(err) {
				c.Redirect(302, podcast.Image)
			} else {
				c.File(localPath)
			}
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	}
}

func GetPodcastItemFileById(c *gin.Context) {
	var searchByIdQuery SearchByIdQuery

	if c.ShouldBindUri(&searchByIdQuery) == nil {

		var podcast db.PodcastItem

		err := db.GetPodcastItemById(searchByIdQuery.Id, &podcast)
		if err == nil {
			if _, err = os.Stat(podcast.DownloadPath); !os.IsNotExist(err) {
				c.Header("Content-Description", "File Transfer")
				c.Header("Content-Transfer-Encoding", "binary")
				c.Header("Content-Disposition", "attachment; filename="+path.Base(podcast.DownloadPath))
				c.Header("Content-Type", GetFileContentType(podcast.DownloadPath))
				c.File(podcast.DownloadPath)
			} else {
				c.Redirect(302, podcast.FileURL)
			}
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	}
}

func GetFileContentType(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return "application/octet-stream"
	}
	defer file.Close()
	buffer := make([]byte, 512)
	if _, err := file.Read(buffer); err != nil {
		return "application/octet-stream"
	}
	return http.DetectContentType(buffer)
}

func MarkPodcastItemAsUnplayed(c *gin.Context) {
	var searchByIdQuery SearchByIdQuery

	if c.ShouldBindUri(&searchByIdQuery) == nil {
		service.SetPodcastItemPlayedStatus(searchByIdQuery.Id, false)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	}
}
func MarkPodcastItemAsPlayed(c *gin.Context) {
	var searchByIdQuery SearchByIdQuery

	if c.ShouldBindUri(&searchByIdQuery) == nil {
		service.SetPodcastItemPlayedStatus(searchByIdQuery.Id, true)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	}
}
func BookmarkPodcastItem(c *gin.Context) {
	var searchByIdQuery SearchByIdQuery

	if c.ShouldBindUri(&searchByIdQuery) == nil {
		service.SetPodcastItemBookmarkStatus(searchByIdQuery.Id, true)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	}
}
func UnbookmarkPodcastItem(c *gin.Context) {
	var searchByIdQuery SearchByIdQuery

	if c.ShouldBindUri(&searchByIdQuery) == nil {
		service.SetPodcastItemBookmarkStatus(searchByIdQuery.Id, false)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	}
}
func PatchPodcastItemById(c *gin.Context) {
	var searchByIdQuery SearchByIdQuery

	if c.ShouldBindUri(&searchByIdQuery) == nil {

		var podcast db.PodcastItem

		err := db.GetPodcastItemById(searchByIdQuery.Id, &podcast)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		var input PatchPodcastItem

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		db.DB.Model(&podcast).Updates(input)
		c.JSON(200, podcast)

	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	}
}

func DownloadPodcastItem(c *gin.Context) {
	var searchByIdQuery SearchByIdQuery

	if c.ShouldBindUri(&searchByIdQuery) == nil {
		go service.DownloadSingleEpisode(searchByIdQuery.Id)
		c.JSON(200, gin.H{})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	}
}
func DeletePodcastItem(c *gin.Context) {
	var searchByIdQuery SearchByIdQuery

	if c.ShouldBindUri(&searchByIdQuery) == nil {

		go service.DeleteEpisodeFile(searchByIdQuery.Id)
		c.JSON(200, gin.H{})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	}
}

func AddPodcast(c *gin.Context) {
	var addPodcastData AddPodcastData
	err := c.ShouldBindJSON(&addPodcastData)
	if err == nil {
		pod, err := service.AddPodcast(addPodcastData.Url)
		if err == nil {
			go service.RefreshEpisodes()
			c.JSON(200, pod)
		} else {
			if v, ok := err.(*model.PodcastAlreadyExistsError); ok {
				c.JSON(409, gin.H{"message": v.Error()})
			} else {
				log.Println(err.Error())
				c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			}
		}
	} else {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	}
}

func GetAllTags(c *gin.Context) {
	tags, err := db.GetAllTags("")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	} else {
		c.JSON(200, tags)
	}

}

func GetTagById(c *gin.Context) {
	var searchByIdQuery SearchByIdQuery
	if c.ShouldBindUri(&searchByIdQuery) == nil {
		tag, err := db.GetTagById(searchByIdQuery.Id)
		if err == nil {
			c.JSON(200, tag)
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	}
}

func getBaseUrl(c *gin.Context) string {
	setting := c.MustGet("setting").(*db.Setting)
	if setting.BaseUrl == "" {
		url := location.Get(c)
		return fmt.Sprintf("%s://%s", url.Scheme, url.Host)
	}
	return setting.BaseUrl
}

func createRss(items []db.PodcastItem, title, description, image string, c *gin.Context) model.RssPodcastData {
	var rssItems []model.RssItem
	url := getBaseUrl(c)
	setting := db.GetOrCreateSetting()
	for _, item := range items {
		var itemGuid string = item.ID
		if setting.PassthroughPodcastGuid && strings.TrimSpace(item.GUID) != "" {
			itemGuid = item.GUID
		}

		rssItem := model.RssItem{
			Title:       item.Title,
			Description: item.Summary,
			Summary:     item.Summary,
			Image: model.RssItemImage{
				Text: item.Title,
				Href: fmt.Sprintf("%s/podcastitems/%s/image", url, item.ID),
			},
			EpisodeType: item.EpisodeType,
			Enclosure: model.RssItemEnclosure{
				URL:    fmt.Sprintf("%s/podcastitems/%s/file", url, item.ID),
				Length: fmt.Sprint(item.FileSize),
				Type:   "audio/mpeg",
			},
			PubDate: item.PubDate.Format("Mon, 02 Jan 2006 15:04:05 -0700"),
			Guid: model.RssItemGuid{
				IsPermaLink: "false",
				Text:        itemGuid,
			},
			Link:     fmt.Sprintf("%s/allTags", url),
			Text:     item.Title,
			Duration: fmt.Sprint(item.Duration),
		}
		rssItems = append(rssItems, rssItem)
	}

	imagePath := fmt.Sprintf("%s/webassets/blank.png", url)
	if image != "" {
		imagePath = image
	}

	return model.RssPodcastData{
		Itunes:  "http://www.itunes.com/dtds/podcast-1.0.dtd",
		Media:   "http://search.yahoo.com/mrss/",
		Version: "2.0",
		Atom:    "http://www.w3.org/2005/Atom",
		Psc:     "https://podlove.org/simple-chapters/",
		Content: "http://purl.org/rss/1.0/modules/content/",
		Channel: model.RssChannel{
			Item:        rssItems,
			Title:       title,
			Description: description,
			Summary:     description,
			Author:      "Podgrab Aggregation",
			Link:        fmt.Sprintf("%s/allTags", url),
			Image:       model.RssItemImage{Text: title, URL: imagePath},
		},
	}
}

func GetRssForPodcastById(c *gin.Context) {
	var searchByIdQuery SearchByIdQuery
	if c.ShouldBindUri(&searchByIdQuery) == nil {
		var podcast db.Podcast
		err := db.GetPodcastById(searchByIdQuery.Id, &podcast)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		}
		var podIds []string
		podIds = append(podIds, searchByIdQuery.Id)
		items := *service.GetAllPodcastItemsByPodcastIds(podIds)

		description := podcast.Summary
		title := podcast.Title

		if err == nil {
			c.XML(200, createRss(items, title, description, podcast.Image, c))
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	}
}
func GetRssForTagById(c *gin.Context) {
	var searchByIdQuery SearchByIdQuery
	if c.ShouldBindUri(&searchByIdQuery) == nil {
		tag, err := db.GetTagById(searchByIdQuery.Id)
		var podIds []string
		for _, pod := range tag.Podcasts {
			podIds = append(podIds, pod.ID)
		}
		items := *service.GetAllPodcastItemsByPodcastIds(podIds)

		description := fmt.Sprintf("Playing episodes with tag : %s", tag.Label)
		title := fmt.Sprintf(" %s | Podgrab", tag.Label)

		if err == nil {
			c.XML(200, createRss(items, title, description, "", c))
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	}
}
func GetRss(c *gin.Context) {
	var items []db.PodcastItem

	if err := db.GetAllPodcastItems(&items); err != nil {
		fmt.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	}

	title := "Podgrab"
	description := "Pograb playlist"

	c.XML(200, createRss(items, title, description, "", c))

}
func DeleteTagById(c *gin.Context) {
	var searchByIdQuery SearchByIdQuery
	if c.ShouldBindUri(&searchByIdQuery) == nil {
		err := service.DeleteTag(searchByIdQuery.Id)
		if err == nil {
			c.JSON(http.StatusNoContent, gin.H{})
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	}
}
func AddTag(c *gin.Context) {
	var addTagData AddTagData
	err := c.ShouldBindJSON(&addTagData)
	if err == nil {
		tag, err := service.AddTag(addTagData.Label, addTagData.Description)
		if err == nil {
			c.JSON(200, tag)
		} else {
			if v, ok := err.(*model.TagAlreadyExistsError); ok {
				c.JSON(409, gin.H{"message": v.Error()})
			} else {
				log.Println(err.Error())
				c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			}
		}
	} else {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	}
}

func AddTagToPodcast(c *gin.Context) {
	var addRemoveTagQuery AddRemoveTagQuery

	if c.ShouldBindUri(&addRemoveTagQuery) == nil {
		err := db.AddTagToPodcast(addRemoveTagQuery.Id, addRemoveTagQuery.TagId)
		if err == nil {
			c.JSON(200, gin.H{})
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	}
}

func RemoveTagFromPodcast(c *gin.Context) {
	var addRemoveTagQuery AddRemoveTagQuery

	if c.ShouldBindUri(&addRemoveTagQuery) == nil {
		err := db.RemoveTagFromPodcast(addRemoveTagQuery.Id, addRemoveTagQuery.TagId)
		if err == nil {
			c.JSON(200, gin.H{})
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	}
}

func UpdateSetting(c *gin.Context) {
	var model SettingModel
	err := c.ShouldBind(&model)

	if err == nil {

		err = service.UpdateSettings(model.DownloadOnAdd, model.InitialDownloadCount,
			model.AutoDownload, model.AppendDateToFileName, model.AppendEpisodeNumberToFileName, model.PassthroughPodcastGuid,
			model.DarkMode, model.DownloadEpisodeImages, model.GenerateNFOFile, model.DontDownloadDeletedFromDisk, model.BaseUrl,
			model.MaxDownloadConcurrency, model.UserAgent,
		)
		if err == nil {
			c.JSON(200, gin.H{"message": "Success"})

		} else {

			c.JSON(http.StatusBadRequest, err)

		}
	} else {
		fmt.Println(err.Error())
		c.JSON(http.StatusBadRequest, err)
	}

}
