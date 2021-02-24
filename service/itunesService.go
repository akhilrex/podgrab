package service

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"

	"github.com/TheHippo/podcastindex"
	"github.com/akhilrex/podgrab/model"
)

type SearchService interface {
	Query(q string) []*model.CommonSearchResultModel
}

type ItunesService struct {
}

const ITUNES_BASE = "https://itunes.apple.com"

func (service ItunesService) Query(q string) []*model.CommonSearchResultModel {
	url := fmt.Sprintf("%s/search?term=%s&entity=podcast", ITUNES_BASE, url.QueryEscape(q))

	body, _ := makeQuery(url)
	var response model.ItunesResponse
	json.Unmarshal(body, &response)

	var toReturn []*model.CommonSearchResultModel

	for _, obj := range response.Results {
		toReturn = append(toReturn, GetSearchFromItunes(obj))
	}

	return toReturn
}

type PodcastIndexService struct {
}

const (
	PODCASTINDEX_KEY    = "LNGTNUAFVL9W2AQKVZ49"
	PODCASTINDEX_SECRET = "H8tq^CZWYmAywbnngTwB$rwQHwMSR8#fJb#Bhgb3"
)

func (service PodcastIndexService) Query(q string) []*model.CommonSearchResultModel {

	c := podcastindex.NewClient(PODCASTINDEX_KEY, PODCASTINDEX_SECRET)
	var toReturn []*model.CommonSearchResultModel
	podcasts, err := c.Search(q)
	if err != nil {
		log.Fatal(err.Error())
		return toReturn
	}

	for _, obj := range podcasts {
		toReturn = append(toReturn, GetSearchFromPodcastIndex(obj))
	}

	return toReturn
}
