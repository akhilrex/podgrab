package db

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func GetPodcastByURL(url string, podcast *Podcast) error {
	result := DB.Preload(clause.Associations).Where(&Podcast{URL: url}).First(&podcast)
	return result.Error
}

func GetPodcastsByURLList(urls []string, podcasts *[]Podcast) error {
	result := DB.Preload(clause.Associations).Where("url in ?", urls).First(&podcasts)
	return result.Error
}
func GetAllPodcasts(podcasts *[]Podcast) error {

	result := DB.Preload("PodcastItems").Find(&podcasts)
	return result.Error
}
func GetAllPodcastItems(podcasts *[]PodcastItem) error {

	result := DB.Preload("Podcast").Order("pub_date desc").Find(&podcasts)
	return result.Error
}
func GetPaginatedPodcastItems(page int, count int, downloadedOnly bool, podcasts *[]PodcastItem, total *int64) error {
	query := DB.Debug().Preload("Podcast")
	if downloadedOnly {
		query = query.Where("download_status=?", Downloaded)
	}

	result := query.Limit(count).Offset((page - 1) * count).Order("pub_date desc").Find(&podcasts)

	DB.Model(&PodcastItem{}).Count(total)

	return result.Error
}
func GetPodcastById(id string, podcast *Podcast) error {

	result := DB.Preload("PodcastItems", func(db *gorm.DB) *gorm.DB {
		return db.Order("podcast_items.pub_date DESC")
	}).First(&podcast, "id=?", id)
	return result.Error
}

func GetPodcastItemById(id string, podcastItem *PodcastItem) error {

	result := DB.Preload(clause.Associations).First(&podcastItem, "id=?", id)
	return result.Error
}
func DeletePodcastItemById(id string) error {

	result := DB.Where("id=?", id).Delete(&PodcastItem{})
	return result.Error
}
func DeletePodcastById(id string) error {

	result := DB.Where("id=?", id).Delete(&Podcast{})
	return result.Error
}

func GetAllPodcastItemsByPodcastId(podcastId string, podcasts *[]PodcastItem) error {

	result := DB.Preload(clause.Associations).Where(&PodcastItem{PodcastID: podcastId}).Find(&podcasts)
	return result.Error
}

func SetAllEpisodesToDownload(podcastId string) error {
	result := DB.Debug().Model(PodcastItem{}).Where(&PodcastItem{PodcastID: podcastId, DownloadStatus: Deleted}).Update("download_status", NotDownloaded)
	return result.Error
}

func GetAllPodcastItemsToBeDownloaded() (*[]PodcastItem, error) {
	var podcastItems []PodcastItem
	result := DB.Debug().Preload(clause.Associations).Where("download_status=?", NotDownloaded).Find(&podcastItems)
	//fmt.Println("To be downloaded : " + string(len(podcastItems)))
	return &podcastItems, result.Error
}
func GetAllPodcastItemsAlreadyDownloaded() (*[]PodcastItem, error) {
	var podcastItems []PodcastItem
	result := DB.Debug().Preload(clause.Associations).Where("download_status=?", Downloaded).Find(&podcastItems)
	return &podcastItems, result.Error
}

func GetPodcastItemsByPodcastIdAndGUIDs(podcastId string, guids []string) (*[]PodcastItem, error) {
	var podcastItems []PodcastItem
	result := DB.Preload(clause.Associations).Where(&PodcastItem{PodcastID: podcastId}).Where("guid IN ?", guids).Find(&podcastItems)
	return &podcastItems, result.Error
}
func GetPodcastItemByPodcastIdAndGUID(podcastId string, guid string, podcastItem *PodcastItem) error {

	result := DB.Preload(clause.Associations).Where(&PodcastItem{PodcastID: podcastId, GUID: guid}).First(&podcastItem)
	return result.Error
}
func GetPodcastByTitleAndAuthor(title string, author string, podcast *Podcast) error {

	result := DB.Preload(clause.Associations).Where(&Podcast{Title: title, Author: author}).First(&podcast)
	return result.Error
}

func CreatePodcast(podcast *Podcast) error {
	tx := DB.Create(&podcast)
	return tx.Error
}

func CreatePodcastItem(podcastItem *PodcastItem) error {
	tx := DB.Omit("Podcast").Create(&podcastItem)
	return tx.Error
}
func UpdatePodcastItem(podcastItem *PodcastItem) error {
	tx := DB.Omit("Podcast").Save(&podcastItem)
	return tx.Error
}
func UpdateSettings(setting *Setting) error {
	tx := DB.Save(&setting)
	return tx.Error
}
func GetOrCreateSetting() *Setting {
	var setting Setting
	result := DB.First(&setting)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		DB.Save(&Setting{})
		DB.First(&setting)
	}
	return &setting
}

func GetLock(name string) *JobLock {
	var jobLock JobLock
	result := DB.Where("name = ?", name).First(&jobLock)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return &JobLock{
			Name: name,
		}
	}
	return &jobLock
}
func Lock(name string, duration int) {
	jobLock := GetLock(name)
	if jobLock == nil {
		jobLock = &JobLock{
			Name: name,
		}
	}
	jobLock.Duration = duration
	jobLock.Date = time.Now()
	if jobLock.ID == "" {
		DB.Create(&jobLock)
	} else {
		DB.Save(&jobLock)
	}
}
func Unlock(name string) {
	jobLock := GetLock(name)
	if jobLock == nil {
		return
	}
	jobLock.Duration = 0
	jobLock.Date = time.Time{}
	DB.Save(&jobLock)
}

func UnlockMissedJobs() {
	var jobLocks *[]JobLock

	result := DB.Where("date != ", time.Time{}).Find(&jobLocks)
	if result.Error != nil {
		return
	}
	for _, job := range *jobLocks {
		var duration time.Duration
		duration = time.Duration(job.Duration)
		d := job.Date.Add(time.Minute * duration)
		if d.Before(time.Now()) {
			Unlock(job.Name)
		}
	}
}
