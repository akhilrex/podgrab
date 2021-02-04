package controllers

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/akhilrex/podgrab/model"
	"github.com/akhilrex/podgrab/service"

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

type Pagination struct {
	Page           int  `uri:"page" query:"page" json:"page" form:"page"`
	Count          int  `uri:"count" query:"count" json:"count" form:"count"`
	DownloadedOnly bool `uri:"downloadedOnly" query:"downloadedOnly" json:"downloadedOnly" form:"downloadedOnly"`
}

type PatchPodcastItem struct {
	IsPlayed bool   `json:"isPlayed" form:"isPlayed" query:"isPlayed"`
	Title    string `form:"title" json:"title" query:"title"`
}

type AddPodcastData struct {
	Url string `binding:"required" form:"url" json:"url"`
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
func DeletePodcastById(c *gin.Context) {
	var searchByIdQuery SearchByIdQuery

	if c.ShouldBindUri(&searchByIdQuery) == nil {

		service.DeletePodcast(searchByIdQuery.Id)
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
	var podcasts []db.PodcastItem
	db.GetAllPodcastItems(&podcasts)
	c.JSON(200, podcasts)
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

func UpdateSetting(c *gin.Context) {
	var model SettingModel
	err := c.ShouldBind(&model)

	if err == nil {

		err = service.UpdateSettings(model.DownloadOnAdd, model.InitialDownloadCount, model.AutoDownload, model.AppendDateToFileName, model.AppendEpisodeNumberToFileName, model.DarkMode)
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
