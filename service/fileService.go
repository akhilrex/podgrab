package service

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"

	stringy "github.com/gobeam/stringy"
)

func Download(link string, episodeTitle string, podcastName string) (string, error) {
	client := httpClient()
	resp, err := client.Get(link)
	if err != nil {
		return "", err
	}

	fileName := getFileName(link, episodeTitle, ".mp3")
	folder := createIfFoldeDoesntExist(podcastName)
	finalPath := path.Join(folder, fileName)
	file, err := os.Create(finalPath)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	_, erra := io.Copy(file, resp.Body)
	//fmt.Println(size)
	defer file.Close()
	if erra != nil {
		return "", erra
	}
	return finalPath, nil

}
func DeleteFile(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return errors.New("File does not exist")
	}
	if err := os.Remove(filePath); err != nil {
		return err
	}
	return nil
}
func httpClient() *http.Client {
	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}

	return &client
}

func createIfFoldeDoesntExist(folder string) string {
	str := stringy.New(folder)
	folder = str.RemoveSpecialCharacter()
	dataPath := os.Getenv("DATA")
	folderPath := path.Join(dataPath, folder)
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		os.MkdirAll(folderPath, 0777)
	}
	return folderPath
}

func getFileName(link string, title string, defaultExtension string) string {
	fileUrl, err := url.Parse(link)
	checkError(err)

	parsed := fileUrl.Path
	ext := filepath.Ext(parsed)

	if len(ext) == 0 {
		ext = defaultExtension
	}
	str := stringy.New(title)
	str = stringy.New(str.RemoveSpecialCharacter())
	return str.KebabCase().Get() + ext

}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
