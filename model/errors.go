package model

import "fmt"

type PodcastAlreadyExistsError struct {
	Url string
}

func (e *PodcastAlreadyExistsError) Error() string {
	return fmt.Sprintf("Podcast with this url already exists")
}
