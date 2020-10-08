package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/akhilrex/podgrab/controllers"
	"github.com/akhilrex/podgrab/db"
	"github.com/akhilrex/podgrab/service"
	"github.com/gin-gonic/gin"
	"github.com/jasonlvhit/gocron"
	_ "github.com/joho/godotenv/autoload"
)

func main() {

	//	os.Remove("./podgrab.db")

	var err error
	db.DB, err = db.Init()
	if err != nil {
		fmt.Println("statuse: ", err)
	} else {
		db.Migrate()
	}
	r := gin.Default()
	dataPath := os.Getenv("DATA")
	//r.Static("/assets", "./assets")
	r.Static("/assets", dataPath)
	funcMap := template.FuncMap{
		"formatDate": func(raw time.Time) string {
			return raw.Format("Jan 2 2006")
		},
	}
	tmpl := template.Must(template.New("main").Funcs(funcMap).ParseGlob("client/*"))

	//r.LoadHTMLGlob("client/*")
	r.SetHTMLTemplate(tmpl)

	r.GET("/podcasts", controllers.AddPodcast)
	r.POST("/podcasts", controllers.GetAllPodcasts)
	r.GET("/podcasts/:id", controllers.GetPodcastById)
	r.GET("/podcasts/:id/items", controllers.GetPodcastItemsByPodcastId)

	r.GET("/podcastitems", controllers.GetAllPodcastItems)
	r.GET("/podcastitems/:id", controllers.GetPodcastItemById)

	r.GET("/", func(c *gin.Context) {
		//var podcasts []db.Podcast
		podcasts := service.GetAllPodcasts()
		c.HTML(http.StatusOK, "index.html", gin.H{"title": "Podgrab", "podcasts": podcasts})
	})
	r.GET("/podcasts/:id/view", func(c *gin.Context) {
		var searchByIdQuery controllers.SearchByIdQuery
		if c.ShouldBindUri(&searchByIdQuery) == nil {

			var podcast db.Podcast

			if err := db.GetPodcastById(searchByIdQuery.Id, &podcast); err == nil {
				c.HTML(http.StatusOK, "podcast.html", gin.H{"title": podcast.Title, "podcast": podcast})
			} else {
				c.JSON(http.StatusBadRequest, err)
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		}

	})
	r.GET("/episodes", func(c *gin.Context) {
		var pagination controllers.Pagination
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
				c.HTML(http.StatusOK, "episodes.html", gin.H{"title": "All Episodes", "podcastItems": podcastItems})
			} else {
				c.JSON(http.StatusBadRequest, err)
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		}

	})
	r.POST(
		"/", func(c *gin.Context) {
			var addPodcastData controllers.AddPodcastData
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

		})

	go assetEnv()
	go intiCron()

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")

}

func intiCron() {
	checkFrequency, err := strconv.Atoi(os.Getenv("CHECK_FREQUENCY"))
	if err != nil {
		checkFrequency = 10
		log.Print(err)
	}
	//gocron.Every(uint64(checkFrequency)).Minutes().Do(service.DownloadMissingEpisodes)
	gocron.Every(uint64(checkFrequency)).Minutes().Do(service.RefreshEpisodes)
	<-gocron.Start()
}

func assetEnv() {
	log.Println("Config Dir: ", os.Getenv("CONFIG"))
	log.Println("Assets Dir: ", os.Getenv("DATA"))
	log.Println("Check Frequency (mins): ", os.Getenv("CHECK_FREQUENCY"))
}
