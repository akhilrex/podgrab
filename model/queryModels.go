package model

import "math"

type Pagination struct {
	Page         int `uri:"page" query:"page" json:"page" form:"page" default:1`
	Count        int `uri:"count" query:"count" json:"count" form:"count" default:20`
	NextPage     int `uri:"nextPage" query:"nextPage" json:"nextPage" form:"nextPage"`
	PreviousPage int `uri:"previousPage" query:"previousPage" json:"previousPage" form:"previousPage"`
	TotalCount   int `uri:"totalCount" query:"totalCount" json:"totalCount" form:"totalCount"`
	TotalPages   int `uri:"totalPages" query:"totalPages" json:"totalPages" form:"totalPages"`
}

type EpisodeSort string

const (
	RELEASE_ASC   EpisodeSort = "release_asc"
	RELEASE_DESC  EpisodeSort = "release_desc"
	DURATION_ASC  EpisodeSort = "duration_asc"
	DURATION_DESC EpisodeSort = "duration_desc"
)

type EpisodesFilter struct {
	Pagination
	DownloadStatus *string     `uri:"downloadStatus" query:"downloadStatus" json:"downloadStatus" form:"downloadStatus"`
	IsPlayed       *string     `uri:"isPlayed" query:"isPlayed" json:"isPlayed" form:"isPlayed"`
	Sorting        EpisodeSort `uri:"sorting" query:"sorting" json:"sorting" form:"sorting"`
	Q              string      `uri:"q" query:"q" json:"q" form:"q"`
	TagIds         []string    `uri:"tagIds" query:"tagIds[]" json:"tagIds" form:"tagIds[]"`
	PodcastIds     []string    `uri:"podcastIds" query:"podcastIds[]" json:"podcastIds" form:"podcastIds[]"`
}

func (filter *EpisodesFilter) VerifyPaginationValues() {
	if filter.Count == 0 {
		filter.Count = 20
	}
	if filter.Page == 0 {
		filter.Page = 1
	}
	if filter.Sorting == "" {
		filter.Sorting = RELEASE_DESC
	}
}

func (filter *EpisodesFilter) SetCounts(totalCount int64) {
	totalPages := int(math.Ceil(float64(totalCount) / float64(filter.Count)))
	nextPage, previousPage := 0, 0
	if filter.Page < totalPages {
		nextPage = filter.Page + 1
	}
	if filter.Page > 1 {
		previousPage = filter.Page - 1
	}
	filter.NextPage = nextPage
	filter.PreviousPage = previousPage
	filter.TotalCount = int(totalCount)
	filter.TotalPages = totalPages
}
