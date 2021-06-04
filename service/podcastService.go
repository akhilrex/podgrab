package service

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/TheHippo/podcastindex"
	"github.com/akhilrex/podgrab/db"
	"github.com/akhilrex/podgrab/model"
	"github.com/akhilrex/podgrab/internal/sanitize"
	"github.com/antchfx/xmlquery"
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
func FetchURL(url string) (model.PodcastData, []byte, error) {
	body, err := makeQuery(url)
	if err != nil {
		return model.PodcastData{}, nil, err
	}
	var response model.PodcastData
	err = xml.Unmarshal(body, &response)
	return response, body, err
}
func GetPodcastById(id string) *db.Podcast {
	var podcast db.Podcast

	db.GetPodcastById(id, &podcast)

	return &podcast
}
func GetPodcastItemById(id string) *db.PodcastItem {
	var podcastItem db.PodcastItem

	db.GetPodcastItemById(id, &podcastItem)

	return &podcastItem
}

func GetAllPodcastItemsByIds(podcastItemIds []string) (*[]db.PodcastItem, error) {
	return db.GetAllPodcastItemsByIds(podcastItemIds)
}
func GetAllPodcastItemsByPodcastIds(podcastIds []string) *[]db.PodcastItem {
	var podcastItems []db.PodcastItem

	db.GetAllPodcastItemsByPodcastIds(podcastIds, &podcastItems)
	return &podcastItems
}

func GetTagsByIds(ids []string) *[]db.Tag {

	tags, _ := db.GetTagsByIds(ids)

	return tags
}
func GetAllPodcasts(sorting string) *[]db.Podcast {
	var podcasts []db.Podcast
	db.GetAllPodcasts(&podcasts, sorting)

	stats, _ := db.GetPodcastEpisodeStats()

	type Key struct {
		PodcastID      string
		DownloadStatus db.DownloadStatus
	}
	countMap := make(map[Key]int)
	sizeMap := make(map[Key]int64)
	for _, stat := range *stats {
		countMap[Key{stat.PodcastID, stat.DownloadStatus}] = stat.Count
		sizeMap[Key{stat.PodcastID, stat.DownloadStatus}] = stat.Size

	}
	var toReturn []db.Podcast
	for _, podcast := range podcasts {
		podcast.DownloadedEpisodesCount = countMap[Key{podcast.ID, db.Downloaded}]
		podcast.DownloadingEpisodesCount = countMap[Key{podcast.ID, db.NotDownloaded}]
		podcast.AllEpisodesCount = podcast.DownloadedEpisodesCount + podcast.DownloadingEpisodesCount + countMap[Key{podcast.ID, db.Deleted}]

		podcast.DownloadedEpisodesSize = sizeMap[Key{podcast.ID, db.Downloaded}]
		podcast.DownloadingEpisodesSize = sizeMap[Key{podcast.ID, db.NotDownloaded}]
		podcast.AllEpisodesSize = podcast.DownloadedEpisodesSize + podcast.DownloadingEpisodesSize + sizeMap[Key{podcast.ID, db.Deleted}]

		toReturn = append(toReturn, podcast)
	}
	return &toReturn
}

func AddOpml(content string) error {
	model, err := ParseOpml(content)
	if err != nil {
		fmt.Println(err.Error())
		return errors.New("Invalid file format")
	}
	var wg sync.WaitGroup
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
	go RefreshEpisodes()
	return nil

}

func ExportOmpl() ([]byte, error) {
	podcasts := GetAllPodcasts("")
	var outlines []model.OpmlOutline
	for _, podcast := range *podcasts {
		toAdd := model.OpmlOutline{
			AttrText: podcast.Summary,
			Type:     "rss",
			XmlUrl:   podcast.URL,
			Title:    podcast.Title,
		}
		outlines = append(outlines, toAdd)
	}

	toExport := model.OpmlExportModel{
		Head: model.OpmlExportHead{
			Title:       "Podgrab Feed Export",
			DateCreated: time.Now(),
		},
		Body: model.OpmlBody{
			Outline: outlines,
		},
		Version: "2.0",
	}

	if data, err := xml.MarshalIndent(toExport, "", "    "); err == nil {
		//	fmt.Println(xml.Header + string(data))
		data = []byte(xml.Header + string(data))
		return data, err
	} else {
		return nil, err
	}
}

func getItunesImageUrl(body []byte) string {
	doc, err := xmlquery.Parse(strings.NewReader(string(body)))
	if err != nil {
		return ""
	}
	channel, err := xmlquery.Query(doc, "//channel")
	if err != nil {
		return ""
	}

	iimage := channel.SelectElement("itunes:image")
	for _, attr := range iimage.Attr {
		if attr.Name.Local == "href" {
			return attr.Value
		}

	}
	return ""

}

func AddPodcast(url string) (db.Podcast, error) {
	var podcast db.Podcast
	err := db.GetPodcastByURL(url, &podcast)
	setting := db.GetOrCreateSetting()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		data, body, err := FetchURL(url)
		if err != nil {
			fmt.Println(err.Error())
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

		if podcast.Image == "" {
			podcast.Image = getItunesImageUrl(body)
		}

		err = db.CreatePodcast(&podcast)
		go DownloadPodcastCoverImage(podcast.Image, podcast.Title)
		if setting.GenerateNFOFile {
			go CreateNfoFile(&podcast)
		}
		return podcast, err
	}

	return podcast, &model.PodcastAlreadyExistsError{Url: url}

}

func AddPodcastItems(podcast *db.Podcast, newPodcast bool) error {
	//fmt.Println("Creating: " + podcast.ID)
	data, _, err := FetchURL(podcast.URL)
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
	var latestDate = time.Time{}
	var itemsAdded = make(map[string]string)
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
			if (pubDate == time.Time{}) {
				//	RFC1123Z    = "Mon, 02 Jan 2006 15:04:05 -0700" // RFC1123 with numeric zone
				modifiedRFC1123Z := "Mon, 2 Jan 2006 15:04:05 -0700"
				pubDate, _ = time.Parse(modifiedRFC1123Z, obj.PubDate)
			}

			if (pubDate == time.Time{}) {
				fmt.Printf("Cant format date : %s", obj.PubDate)
			}

			if latestDate.Before(pubDate) {
				latestDate = pubDate
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

			if newPodcast && !setting.DownloadOnAdd {
				downloadStatus = db.Deleted
			}

			summary := strip.StripTags(obj.Summary)
			if summary == "" {
				summary = strip.StripTags(obj.Description)
			}

			podcastItem = db.PodcastItem{
				PodcastID:      podcast.ID,
				Title:          obj.Title,
				Summary:        summary,
				EpisodeType:    obj.EpisodeType,
				Duration:       duration,
				PubDate:        pubDate,
				FileURL:        obj.Enclosure.URL,
				GUID:           obj.Guid.Text,
				Image:          obj.Image.Href,
				DownloadStatus: downloadStatus,
			}
			db.CreatePodcastItem(&podcastItem)
			itemsAdded[podcastItem.ID] = podcastItem.FileURL
		}
	}
	if (latestDate != time.Time{}) {
		db.UpdateLastEpisodeDateForPodcast(podcast.ID, latestDate)
	}
	//go updateSizeFromUrl(itemsAdded)
	return err
}

func updateSizeFromUrl(itemUrlMap map[string]string) {

	for id, url := range itemUrlMap {
		size, err := GetFileSizeFromUrl(url)
		if err != nil {
			size = 1
		}

		db.UpdatePodcastItemFileSize(id, size)
	}

}

func UpdateAllFileSizes() {
	items, err := db.GetAllPodcastItemsWithoutSize()
	if err != nil {
		return
	}
	for _, item := range *items {
		var size int64 = 1
		if item.DownloadStatus == db.Downloaded {
			size, _ = GetFileSize(item.DownloadPath)
		} else {
			size, _ = GetFileSizeFromUrl(item.FileURL)
		}
		db.UpdatePodcastItemFileSize(item.ID, size)
	}
}

func SetPodcastItemAsQueuedForDownload(id string) error {
	var podcastItem db.PodcastItem
	err := db.GetPodcastItemById(id, &podcastItem)
	if err != nil {
		return err
	}
	podcastItem.DownloadStatus = db.NotDownloaded

	return db.UpdatePodcastItem(&podcastItem)
}

func DownloadMissingImages() error {
	setting := db.GetOrCreateSetting()
	if !setting.DownloadEpisodeImages {
		fmt.Println("No Need To Download Images")
		return nil
	}
	items, err := db.GetAllPodcastItemsWithoutImage()
	if err != nil {
		return err
	}
	for _, item := range *items {
		downloadImageLocally(item.ID)
	}
	return nil
}

func downloadImageLocally(podcastItemId string) error {
	var podcastItem db.PodcastItem
	err := db.GetPodcastItemById(podcastItemId, &podcastItem)
	if err != nil {
		return err
	}

	path, err := DownloadImage(podcastItem.Image, podcastItem.ID, podcastItem.Podcast.Title)
	if err != nil {
		return err
	}

	podcastItem.LocalImage = path

	return db.UpdatePodcastItem(&podcastItem)
}

func SetPodcastItemBookmarkStatus(id string, bookmark bool) error {
	var podcastItem db.PodcastItem
	err := db.GetPodcastItemById(id, &podcastItem)
	if err != nil {
		return err
	}
	if bookmark {
		podcastItem.BookmarkDate = time.Now()
	} else {
		podcastItem.BookmarkDate = time.Time{}
	}
	return db.UpdatePodcastItem(&podcastItem)
}

func SetPodcastItemAsDownloaded(id string, location string) error {
	var podcastItem db.PodcastItem

	err := db.GetPodcastItemById(id, &podcastItem)
	if err != nil {
		fmt.Println("Location", err.Error())
		return err
	}

	size, err := GetFileSize(location)
	if err == nil {
		podcastItem.FileSize = size
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
	var podcast db.Podcast
	err := db.GetPodcastById(podcastId, &podcast)
	if err != nil {
		return err
	}
	AddPodcastItems(&podcast, false)
	return db.SetAllEpisodesToDownload(podcastId)
}

var format_re = regexp.MustCompile("%%|%([^%]+)%")

var format_map = map[string]interface{}{
	"%ShowTitle%": func(item *db.PodcastItem) string {
		return item.Podcast.Title
	},
	"%EpisodeTitle%": func(item *db.PodcastItem) string {
		return item.Title
	},
	"%EpisodeNumber%": func(item *db.PodcastItem) string {
		seq, err := db.GetEpisodeNumber(item.ID, item.PodcastID)
		if err == nil {
			return strconv.Itoa(seq)
		}
		return "0"
	},
	"%EpisodeDate%": func(item *db.PodcastItem) string {
		return item.PubDate.Format("2006-01-02")
	},
	"%YYYY%": func(item *db.PodcastItem) string {
		return item.PubDate.Format("2006")
	},
	"%mm%": func(item *db.PodcastItem) string {
		return item.PubDate.Format("01")
	},
	"%dd%": func(item *db.PodcastItem) string {
		return item.PubDate.Format("02")
	},
	"%%": func(item *db.PodcastItem) string {
		return "%"
	},
}

func FormatFileName(item *db.PodcastItem, formatString string) string {
	var matchedTokens = format_re.FindAllStringIndex(formatString, -1)
	var previousTokenEndingIndex = 0
	var formattedFileName = ""
	for _,t := range matchedTokens {
		if previousTokenEndingIndex != t[0] {
			formattedFileName += formatString[previousTokenEndingIndex:t[0]]
		}
		token := formatString[t[0]:t[1]]
		tokenFunction := format_map[token]
		if nil != tokenFunction {
			token = tokenFunction.(func(item *db.PodcastItem)string)(item)
			token = sanitize.Name(token)
		}
		formattedFileName += token
		previousTokenEndingIndex = t[1]
	}
	if previousTokenEndingIndex < len(formatString) {
		formattedFileName += formatString[previousTokenEndingIndex:]
	}
	return formattedFileName
}

func DownloadMissingEpisodes() error {
	const JOB_NAME = "DownloadMissingEpisodes"
	lock := db.GetLock(JOB_NAME)
	if lock.IsLocked() {
		fmt.Println(JOB_NAME + " is locked")
		return nil
	}
	db.Lock(JOB_NAME, 120)
	setting := db.GetOrCreateSetting()

	data, err := db.GetAllPodcastItemsToBeDownloaded()

	fmt.Println("Processing episodes: ", strconv.Itoa(len(*data)))
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	for index, item := range *data {
		wg.Add(1)
		go func(item db.PodcastItem, setting db.Setting) {
			defer wg.Done()
			podcastFileName := FormatFileName(&item, setting.FileNameFormat)
			url, _ := Download(item.FileURL, item.Title, item.Podcast.Title, podcastFileName)
			SetPodcastItemAsDownloaded(item.ID, url)
		}(item, *setting)

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
		fmt.Println(err.Error())
		return err
	}

	if podcastItem.LocalImage != "" {
		go DeleteFile(podcastItem.LocalImage)
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

	setting := db.GetOrCreateSetting()
	SetPodcastItemAsQueuedForDownload(podcastItemId)

	podcastFileName := FormatFileName(&podcastItem, setting.FileNameFormat)
	url, err := Download(podcastItem.FileURL, podcastItem.Title, podcastItem.Podcast.Title, podcastFileName)

	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	err = SetPodcastItemAsDownloaded(podcastItem.ID, url)

	if setting.DownloadEpisodeImages {
		downloadImageLocally(podcastItem.ID)
	}
	return err
}

func RefreshEpisodes() error {
	var data []db.Podcast
	err := db.GetAllPodcasts(&data, "")

	if err != nil {
		return err
	}
	for _, item := range data {
		isNewPodcast := item.LastEpisode == nil
		if isNewPodcast {
			fmt.Println(item.Title)
			db.ForceSetLastEpisodeDate(item.ID)
		}
		AddPodcastItems(&item, isNewPodcast)
	}
	//	setting := db.GetOrCreateSetting()

	go DownloadMissingEpisodes()

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
		if item.LocalImage != "" {
			DeleteFile(item.LocalImage)
		}
		SetPodcastItemAsNotDownloaded(item.ID, db.Deleted)

	}
	return nil

}
func DeletePodcast(id string, deleteFiles bool) error {
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
		if deleteFiles {
			DeleteFile(item.DownloadPath)
			if item.LocalImage != "" {
				DeleteFile(item.LocalImage)
			}

		}
		db.DeletePodcastItemById(item.ID)

	}
	err = db.DeletePodcastById(id)
	if err != nil {
		return err
	}
	return nil

}
func DeleteTag(id string) error {
	db.UntagAllByTagId(id)
	err := db.DeleteTagById(id)
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
func GetSearchFromPodcastIndex(pod *podcastindex.Podcast) *model.CommonSearchResultModel {
	p := new(model.CommonSearchResultModel)
	p.URL = pod.URL
	p.Image = pod.Image
	p.Title = pod.Title
	p.Description = pod.Description

	if pod.Categories != nil {
		values := make([]string, 0, len(pod.Categories))
		for _, val := range pod.Categories {
			values = append(values, val)
		}
		p.Categories = values
	}

	return p
}

func UpdateSettings(downloadOnAdd bool, initialDownloadCount int, autoDownload bool, fileNameFormat string, darkMode bool, downloadEpisodeImages bool, generateNFOFile bool) error {
	setting := db.GetOrCreateSetting()

	setting.AutoDownload = autoDownload
	setting.DownloadOnAdd = downloadOnAdd
	setting.InitialDownloadCount = initialDownloadCount
	setting.FileNameFormat = fileNameFormat
	setting.DarkMode = darkMode
	setting.DownloadEpisodeImages = downloadEpisodeImages
	setting.GenerateNFOFile = generateNFOFile

	return db.UpdateSettings(setting)
}

func UnlockMissedJobs() {
	db.UnlockMissedJobs()
}

func AddTag(label, description string) (db.Tag, error) {

	tag, err := db.GetTagByLabel(label)

	if errors.Is(err, gorm.ErrRecordNotFound) {

		tag := db.Tag{
			Label:       label,
			Description: description,
		}

		err = db.CreateTag(&tag)
		return tag, err
	}

	return *tag, &model.TagAlreadyExistsError{Label: label}

}
