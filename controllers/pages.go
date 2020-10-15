package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/akhilrex/podgrab/db"
	"github.com/akhilrex/podgrab/service"
	"github.com/gin-gonic/gin"
)

type SearchGPodderData struct {
	Q string `binding:"required" form:"q" json:"q" query:"q"`
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
			c.HTML(http.StatusOK, "podcast.html", gin.H{"title": podcast.Title, "podcast": podcast, "setting": setting})
		} else {
			c.JSON(http.StatusBadRequest, err)
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
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

		if err := db.GetPaginatedPodcastItems(page, count, &podcastItems); err == nil {
			setting := c.MustGet("setting").(*db.Setting)
			c.HTML(http.StatusOK, "episodes.html", gin.H{"title": "All Episodes", "podcastItems": podcastItems, "setting": setting})
		} else {
			c.JSON(http.StatusBadRequest, err)
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	}

}

func Search(c *gin.Context) {
	var searchQuery SearchGPodderData
	if c.ShouldBindQuery(&searchQuery) == nil {
		data := service.Query(searchQuery.Q)
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
