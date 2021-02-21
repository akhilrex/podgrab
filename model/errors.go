package model

import "fmt"

type PodcastAlreadyExistsError struct {
	Url string
}

func (e *PodcastAlreadyExistsError) Error() string {
	return fmt.Sprintf("Podcast with this url already exists")
}

type TagAlreadyExistsError struct {
	Label string
}

func (e *TagAlreadyExistsError) Error() string {
	return fmt.Sprintf("Tag with this label already exists : " + e.Label)
}
