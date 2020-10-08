package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

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
	}else{
		db.Migrate()
	}
	r := gin.Default()
	dataPath := os.Getenv("DATA")
	//r.Static("/assets", "./assets")
	r.Static("/assets", dataPath)
	r.LoadHTMLGlob("client/*")

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
