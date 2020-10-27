package service

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/akhilrex/podgrab/model"
)

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
