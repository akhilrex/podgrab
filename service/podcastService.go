package service

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/akhilrex/podgrab/db"
	"github.com/akhilrex/podgrab/model"
	strip "github.com/grokify/html-strip-tags-go"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var Logger *zap.SugaredLogger

func init() {
	zapper, _ := zap.NewProduction()
	Logger = zapper.Sugar()
	defer zapper.Sync()
}

func ParseOpml(content string) (model.OpmlModel, error) {
	var response model.OpmlModel
	err := xml.Unmarshal([]byte(content), &response)
	return response, err
}

//FetchURL is
func FetchURL(url string) (model.PodcastData, error) {
	body, err := makeQuery(url)
	if err != nil {
		return model.PodcastData{}, err
	}
	var response model.PodcastData
	err = xml.Unmarshal(body, &response)
	return response, err
}
func GetAllPodcasts() *[]db.Podcast {
	var podcasts []db.Podcast
	db.GetAllPodcasts(&podcasts)
	return &podcasts
}

func AddOpml(content string) error {
	model, err := ParseOpml(content)
	if err != nil {
		return errors.New("Invalid file format")
	}
	var wg sync.WaitGroup
	setting := db.GetOrCreateSetting()
	for _, outline := range model.Body.Outline {
		if outline.XmlUrl != "" {
			wg.Add(1)
			go func(url string) {
				defer wg.Done()
				AddPodcast(url)

			}(outline.XmlUrl)
		}

		for _, innerOutline := range outline.Outline {
			if innerOutline.XmlUrl != "" {
				wg.Add(1)
				go func(url string) {
					defer wg.Done()
					AddPodcast(url)
				}(innerOutline.XmlUrl)
			}
		}
	}
	wg.Wait()
	if setting.DownloadOnAdd {
		go RefreshEpisodes()
	}
	return nil

}

func ExportOmpl() (model.OpmlModel, error) {
	podcasts := GetAllPodcasts()
	var outlines []model.OpmlOutline
	for _, podcast := range *podcasts {
		toAdd := model.OpmlOutline{
			AttrText: podcast.Title,
			Type:     "rss",
			XmlUrl:   podcast.URL,
		}
		outlines = append(outlines, toAdd)
	}

	toExport := model.OpmlModel{
		Head: model.OpmlHead{
			Title: "Podgrab Feed Export",
		},
		Body: model.OpmlBody{
			Outline: outlines,
		},
		Version: "1.0",
	}

	data, err := xml.Marshal(toExport)
	//return string(data), err
	fmt.Println(string(data))
	return toExport, err
}
func AddPodcast(url string) (db.Podcast, error) {
	var podcast db.Podcast
	err := db.GetPodcastByURL(url, &podcast)
	fmt.Println(url)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		data, err := FetchURL(url)
		if err != nil {
			fmt.Println("Error")
			Logger.Errorw("Error adding podcast", err)
			return db.Podcast{}, err
		}

		podcast := db.Podcast{
			Title:   data.Channel.Title,
			Summary: strip.StripTags(data.Channel.Summary),
			Author:  data.Channel.Author,
			Image:   data.Channel.Image.URL,
			URL:     url,
		}
		err = db.CreatePodcast(&podcast)
		return podcast, err
	}
	return podcast, &model.PodcastAlreadyExistsError{Url: url}

}

func AddPodcastItems(podcast *db.Podcast) error {
	//fmt.Println("Creating: " + podcast.ID)
	data, err := FetchURL(podcast.URL)
	if err != nil {
		//log.Fatal(err)
		return err
	}
	setting := db.GetOrCreateSetting()
	limit := setting.InitialDownloadCount
	// if len(data.Channel.Item) < limit {
	// 	limit = len(data.Channel.Item)
	// }
	var allGuids []string
	for i := 0; i < len(data.Channel.Item); i++ {
		obj := data.Channel.Item[i]
		allGuids = append(allGuids, obj.Guid.Text)
	}

	existingItems, err := db.GetPodcastItemsByPodcastIdAndGUIDs(podcast.ID, allGuids)
	keyMap := make(map[string]int)

	for _, item := range *existingItems {
		keyMap[item.GUID] = 1
	}

	for i := 0; i < len(data.Channel.Item); i++ {
		obj := data.Channel.Item[i]
		var podcastItem db.PodcastItem
		_, keyExists := keyMap[obj.Guid.Text]
		if !keyExists {
			duration, _ := strconv.Atoi(obj.Duration)
			pubDate, _ := time.Parse(time.RFC1123Z, obj.PubDate)
			if (pubDate == time.Time{}) {
				pubDate, _ = time.Parse(time.RFC1123, obj.PubDate)
			}
			if (pubDate == time.Time{}) {
				//	RFC1123     = "Mon, 02 Jan 2006 15:04:05 MST"
				modifiedRFC1123 := "Mon, 2 Jan 2006 15:04:05 MST"
				pubDate, _ = time.Parse(modifiedRFC1123, obj.PubDate)
			}

			var downloadStatus db.DownloadStatus
			if setting.AutoDownload {
				if i < limit {
					downloadStatus = db.NotDownloaded
				} else {
					downloadStatus = db.Deleted
				}
			} else {
				downloadStatus = db.Deleted
			}
			podcastItem = db.PodcastItem{
				PodcastID:      podcast.ID,
				Title:          obj.Title,
				Summary:        strip.StripTags(obj.Summary),
				EpisodeType:    obj.EpisodeType,
				Duration:       duration,
				PubDate:        pubDate,
				FileURL:        obj.Enclosure.URL,
				GUID:           obj.Guid.Text,
				Image:          obj.Image.Href,
				DownloadStatus: downloadStatus,
			}
			db.CreatePodcastItem(&podcastItem)
		}
	}
	return err
}

func SetPodcastItemAsDownloaded(id string, location string) error {
	var podcastItem db.PodcastItem
	err := db.GetPodcastItemById(id, &podcastItem)
	if err != nil {
		return err
	}
	podcastItem.DownloadDate = time.Now()
	podcastItem.DownloadPath = location
	podcastItem.DownloadStatus = db.Downloaded

	return db.UpdatePodcastItem(&podcastItem)
}
func SetPodcastItemAsNotDownloaded(id string, downloadStatus db.DownloadStatus) error {
	var podcastItem db.PodcastItem
	err := db.GetPodcastItemById(id, &podcastItem)
	if err != nil {
		return err
	}
	podcastItem.DownloadDate = time.Time{}
	podcastItem.DownloadPath = ""
	podcastItem.DownloadStatus = downloadStatus

	return db.UpdatePodcastItem(&podcastItem)
}

func SetPodcastItemPlayedStatus(id string, isPlayed bool) error {
	var podcastItem db.PodcastItem
	err := db.GetPodcastItemById(id, &podcastItem)
	if err != nil {
		return err
	}
	podcastItem.IsPlayed = isPlayed
	return db.UpdatePodcastItem(&podcastItem)
}
func SetAllEpisodesToDownload(podcastId string) error {
	return db.SetAllEpisodesToDownload(podcastId)
}
func DownloadMissingEpisodes() error {
	const JOB_NAME = "DownloadMissingEpisodes"
	lock := db.GetLock(JOB_NAME)
	if lock.IsLocked() {
		fmt.Println(JOB_NAME + " is locked")
		return nil
	}
	db.Lock(JOB_NAME, 120)

	data, err := db.GetAllPodcastItemsToBeDownloaded()

	fmt.Println("Processing episodes: ", strconv.Itoa(len(*data)))
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	for index, item := range *data {
		wg.Add(1)
		go func(item db.PodcastItem) {
			defer wg.Done()
			url, _ := Download(item.FileURL, item.Title, item.Podcast.Title)
			SetPodcastItemAsDownloaded(item.ID, url)
		}(item)

		if index%5 == 0 {
			wg.Wait()
		}
	}
	wg.Wait()
	db.Unlock(JOB_NAME)
	return nil
}
func CheckMissingFiles() error {
	data, err := db.GetAllPodcastItemsAlreadyDownloaded()

	//fmt.Println("Processing episodes: ", strconv.Itoa(len(*data)))
	if err != nil {
		return err
	}
	for _, item := range *data {
		fileExists := FileExists(item.DownloadPath)
		if !fileExists {
			SetPodcastItemAsNotDownloaded(item.ID, db.NotDownloaded)
		}
	}
	return nil
}

func DeleteEpisodeFile(podcastItemId string) error {
	var podcastItem db.PodcastItem
	err := db.GetPodcastItemById(podcastItemId, &podcastItem)

	//fmt.Println("Processing episodes: ", strconv.Itoa(len(*data)))
	if err != nil {
		return err
	}

	err = DeleteFile(podcastItem.DownloadPath)

	if err != nil && !os.IsNotExist(err) {
		return err
	}

	return SetPodcastItemAsNotDownloaded(podcastItem.ID, db.Deleted)
}
func DownloadSingleEpisode(podcastItemId string) error {
	var podcastItem db.PodcastItem
	err := db.GetPodcastItemById(podcastItemId, &podcastItem)

	//fmt.Println("Processing episodes: ", strconv.Itoa(len(*data)))
	if err != nil {
		return err
	}

	url, err := Download(podcastItem.FileURL, podcastItem.Title, podcastItem.Podcast.Title)
	if err != nil {
		return err
	}
	return SetPodcastItemAsDownloaded(podcastItem.ID, url)
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
	setting := db.GetOrCreateSetting()
	if setting.AutoDownload {
		go DownloadMissingEpisodes()
	}
	return nil
}

func DeletePodcastEpisodes(id string) error {
	var podcast db.Podcast

	err := db.GetPodcastById(id, &podcast)
	if err != nil {
		return err
	}
	var podcastItems []db.PodcastItem

	err = db.GetAllPodcastItemsByPodcastId(id, &podcastItems)
	if err != nil {
		return err
	}
	for _, item := range podcastItems {
		DeleteFile(item.DownloadPath)
		SetPodcastItemAsNotDownloaded(item.ID, db.Deleted)

	}
	return nil

}
func DeletePodcast(id string) error {
	var podcast db.Podcast

	err := db.GetPodcastById(id, &podcast)
	if err != nil {
		return err
	}
	var podcastItems []db.PodcastItem

	err = db.GetAllPodcastItemsByPodcastId(id, &podcastItems)
	if err != nil {
		return err
	}
	for _, item := range podcastItems {
		DeleteFile(item.DownloadPath)
		db.DeletePodcastItemById(item.ID)

	}
	err = db.DeletePodcastById(id)
	if err != nil {
		return err
	}
	return nil

}

func makeQuery(url string) ([]byte, error) {
	//link := "https://www.goodreads.com/search/index.xml?q=Good%27s+Omens&key=" + "jCmNlIXjz29GoB8wYsrd0w"
	//link := "https://www.goodreads.com/search/index.xml?key=jCmNlIXjz29GoB8wYsrd0w&q=Ender%27s+Game"
	fmt.Println(url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	fmt.Println("Response status:", resp.Status)
	body, err := ioutil.ReadAll(resp.Body)

	return body, nil

}
func GetSearchFromGpodder(pod model.GPodcast) *model.CommonSearchResultModel {
	p := new(model.CommonSearchResultModel)
	p.URL = pod.URL
	p.Image = pod.LogoURL
	p.Title = pod.Title
	p.Description = pod.Description
	return p
}
func GetSearchFromItunes(pod model.ItunesSingleResult) *model.CommonSearchResultModel {
	p := new(model.CommonSearchResultModel)
	p.URL = pod.FeedURL
	p.Image = pod.ArtworkURL600
	p.Title = pod.TrackName

	return p
}

func UpdateSettings(downloadOnAdd bool, initialDownloadCount int, autoDownload bool) error {
	setting := db.GetOrCreateSetting()

	setting.AutoDownload = autoDownload
	setting.DownloadOnAdd = downloadOnAdd
	setting.InitialDownloadCount = initialDownloadCount

	return db.UpdateSettings(setting)
}

func UnlockMissedJobs() {
	db.UnlockMissedJobs()
}
