package service

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/akhilrex/podgrab/model"
)

// type GoodReadsService struct {
// }

const BASE = "https://gpodder.net"

func Query(q string) []*model.CommonSearchResultModel {
	url := fmt.Sprintf("%s/search.json?q=%s", BASE, url.QueryEscape(q))

	body, _ := makeQuery(url)
	var response []model.GPodcast
	json.Unmarshal(body, &response)

	var toReturn []*model.CommonSearchResultModel

	for _, obj := range response {
		toReturn = append(toReturn, GetSearchFromGpodder(obj))
	}

	return toReturn
}
func ByTag(tag string, count int) []model.GPodcast {
	url := fmt.Sprintf("%s/api/2/tag/%s/%d.json", BASE, url.QueryEscape(tag), count)

	body, _ := makeQuery(url)
	var response []model.GPodcast
	json.Unmarshal(body, &response)
	return response
}
func Top(count int) []model.GPodcast {
	url := fmt.Sprintf("%s/toplist/%d.json", BASE, count)

	body, _ := makeQuery(url)
	var response []model.GPodcast
	json.Unmarshal(body, &response)
	return response
}
func Tags(count int) []model.GPodcastTag {
	url := fmt.Sprintf("%s/api/2/tags/%d.json", BASE, count)

	body, _ := makeQuery(url)
	var response []model.GPodcastTag
	json.Unmarshal(body, &response)
	return response
}
