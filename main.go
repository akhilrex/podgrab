package main

import (
	"fmt"
	"html/template"
	"log"
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
	r.Static("/webassets", "./webassets")
	r.Static("/assets", dataPath)
	r.Use(setupSettings())

	funcMap := template.FuncMap{
		"formatDate": func(raw time.Time) string {
			return raw.Format("Jan 2 2006")
		},
		"latestEpisodeDate": func(podcastItems []db.PodcastItem) string {
			var latest time.Time
			for _, item := range podcastItems {
				if item.PubDate.After(latest) {
					latest = item.PubDate
				}
			}
			return latest.Format("Jan 2 2006")
		},
		"formatDuration": func(total int) string {
			if total <= 0 {
				return ""
			}
			mins := total / 60
			secs := total % 60
			hrs := 0
			if mins >= 60 {
				hrs = mins / 60
				mins = mins % 60
			}
			if hrs > 0 {
				return fmt.Sprintf("%02d:%02d:%02d", hrs, mins, secs)
			}
			return fmt.Sprintf("%02d:%02d", mins, secs)
		},
	}
	tmpl := template.Must(template.New("main").Funcs(funcMap).ParseGlob("client/*"))

	//r.LoadHTMLGlob("client/*")
	r.SetHTMLTemplate(tmpl)

	r.POST("/podcasts", controllers.AddPodcast)
	r.GET("/podcasts", controllers.GetAllPodcasts)
	r.GET("/podcasts/:id", controllers.GetPodcastById)
	r.DELETE("/podcasts/:id", controllers.DeletePodcastById)
	r.GET("/podcasts/:id/items", controllers.GetPodcastItemsByPodcastId)

	r.GET("/podcastitems", controllers.GetAllPodcastItems)
	r.GET("/podcastitems/:id", controllers.GetPodcastItemById)
	r.GET("/podcastitems/:id/download", controllers.DownloadPodcastItem)

	r.GET("/add", controllers.AddPage)
	r.GET("/search", controllers.Search)
	r.GET("/", controllers.HomePage)
	r.GET("/podcasts/:id/view", controllers.PodcastPage)
	r.GET("/episodes", controllers.AllEpisodesPage)

	go assetEnv()
	go intiCron()

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")

}
func setupSettings() gin.HandlerFunc {
	return func(c *gin.Context) {

		setting := db.GetOrCreateSetting()
		c.Set("setting", setting)
		c.Next()
	}
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
