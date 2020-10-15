package model

type GPodcast struct {
	URL                 string `json:"url"`
	Title               string `json:"title"`
	Author              string `json:"author"`
	Description         string `json:"description"`
	Subscribers         int    `json:"subscribers"`
	SubscribersLastWeek int    `json:"subscribers_last_week"`
	LogoURL             string `json:"logo_url"`
	ScaledLogoURL       string `json:"scaled_logo_url"`
	Website             string `json:"website"`
	MygpoLink           string `json:"mygpo_link"`
	AlreadySaved        bool   `json:"already_saved"`
}

type GPodcastTag struct {
	Tag   string `json:"tag"`
	Title string `json:"title"`
	Usage int    `json:"usage"`
}
