package controllers

import (
	"fmt"
	"net/http"

	"github.com/akhilrex/podgrab/service"

	"github.com/akhilrex/podgrab/db"
	"github.com/gin-gonic/gin"
)

type SearchQuery struct {
	Q    string `binding:"required" form:"q"`
	Type string `form:"type"`
}

type SearchByIdQuery struct {
	Id string `binding:"required" uri:"id" json:"id" form:"id"`
}

type AddPodcastData struct {
	url string `binding:"required" form:"url" json:"url"`
}

func GetAllPodcasts(c *gin.Context) {
	var podcasts []db.Podcast
	db.GetAllPodcasts(&podcasts)
	c.JSON(200, podcasts)
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

func AddPodcast(c *gin.Context) {
	var addPodcastData AddPodcastData
	err := c.ShouldBindJSON(&addPodcastData)
	if err == nil {

		service.AddPodcast(addPodcastData.url)
		//	fmt.Println(time.Unix(addPodcastData.StartDate, 0))
		c.JSON(200, addPodcastData)
	} else {
		fmt.Println(err.Error())
		c.JSON(http.StatusBadRequest, err)
	}
}
