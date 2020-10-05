package service

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/akhilrex/podgrab/db"
	"gorm.io/gorm"
)

//Fetch is
func FetchURL(url string) (PodcastData, error) {
	body := makeQuery(url)
	var response PodcastData
	err := xml.Unmarshal(body, &response)
	return response, err
}
func GetAllPodcasts() *[]db.Podcast {
	var podcasts []db.Podcast
	db.GetAllPodcasts(&podcasts)
	return &podcasts
}
func AddPodcast(url string) (db.Podcast, error) {

	data, err := FetchURL(url)
	if err != nil {
		fmt.Println("Error")
		log.Fatal(err)
		return db.Podcast{}, err
	}
	var podcast db.Podcast
	err = db.GetPodcastByTitleAndAuthor(data.Channel.Title, data.Channel.Author, &podcast)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		podcast := db.Podcast{
			Title:   data.Channel.Title,
			Summary: data.Channel.Summary,
			Author:  data.Channel.Author,
			Image:   data.Channel.Image.URL,
			URL:     url,
		}
		err = db.CreatePodcast(&podcast)
		return podcast, err
	}
	return podcast, err

}

func AddPodcastItems(podcast *db.Podcast) error {
	fmt.Println("Creating: " + podcast.ID)
	data, err := FetchURL(podcast.URL)
	if err != nil {
		log.Fatal(err)
		return err
	}

	for i := 0; i < 5; i++ {
		obj := data.Channel.Item[i]
		var podcastItem db.PodcastItem
		err := db.GetPodcastItemByPodcastIdAndGUID(podcast.ID, obj.Guid.Text, &podcastItem)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			duration, _ := strconv.Atoi(obj.Duration)
			pubDate, _ := time.Parse(time.RFC1123Z, obj.PubDate)
			podcastItem = db.PodcastItem{
				PodcastID:   podcast.ID,
				Title:       obj.Title,
				Summary:     obj.Summary,
				EpisodeType: obj.EpisodeType,
				Duration:    duration,
				PubDate:     pubDate,
				FileURL:     obj.Enclosure.URL,
				GUID:        obj.Guid.Text,
			}
			db.CreatePodcastItem(&podcastItem)
		}
	}
	return err
}

func SetPodcastItemAsDownloaded(id string, location string) {
	var podcastItem db.PodcastItem
	db.GetPodcastItemById(id, &podcastItem)

	podcastItem.DownloadDate = time.Now()
	podcastItem.DownloadPath = location

	db.UpdatePodcastItem(&podcastItem)
}

func DownloadMissingEpisodes() error {
	data, err := db.GetAllPodcastItemsToBeDownloaded()

	fmt.Println("Processing episodes: %d", len(*data))
	if err != nil {
		return err
	}
	for _, item := range *data {
		url, _ := Download(item.FileURL, item.Title, item.Podcast.Title)
		SetPodcastItemAsDownloaded(item.ID, url)
	}
	return nil
}

func RefreshEpisodes() error {
	var data []db.Podcast
	err := db.GetAllPodcasts(&data)

	if err != nil {
		return err
	}
	for _, item := range data {
		AddPodcastItems(&item)

	}
	return nil
}

func makeQuery(url string) []byte {
	//link := "https://www.goodreads.com/search/index.xml?q=Good%27s+Omens&key=" + "jCmNlIXjz29GoB8wYsrd0w"
	//link := "https://www.goodreads.com/search/index.xml?key=jCmNlIXjz29GoB8wYsrd0w&q=Ender%27s+Game"
	//fmt.Println(url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()
	fmt.Println("Response status:", resp.Status)
	body, err := ioutil.ReadAll(resp.Body)

	return body

}
