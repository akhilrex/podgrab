package db

import (
	"time"

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
func GetPaginatedPodcastItems(page int, count int, podcasts *[]PodcastItem) error {
	result := DB.Debug().Preload("Podcast").Limit(count).Offset((page - 1) * count).Order("pub_date desc").Find(&podcasts)
	return result.Error
}
func GetPodcastById(id string, podcast *Podcast) error {

	result := DB.Preload(clause.Associations).First(&podcast, "id=?", id)
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

func GetAllPodcastItemsToBeDownloaded() (*[]PodcastItem, error) {
	var podcastItems []PodcastItem
	result := DB.Debug().Preload(clause.Associations).Where("download_date=?", time.Time{}).Find(&podcastItems)
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
