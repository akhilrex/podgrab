package service

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	stringy "github.com/gobeam/stringy"
)

func Download(link string, episodeTitle string, podcastName string) (string, error) {
	if link == "" {
		return "", errors.New("Download path empty")
	}
	client := httpClient()
	resp, err := client.Get(link)
	if err != nil {
		Logger.Errorw("Error getting response: "+link, err)
		return "", err
	}

	fileName := getFileName(link, episodeTitle, ".mp3")
	folder := createIfFoldeDoesntExist(podcastName)
	finalPath := path.Join(folder, fileName)
	file, err := os.Create(finalPath)
	if err != nil {
		Logger.Errorw("Error creating file"+link, err)
		return "", err
	}
	defer resp.Body.Close()
	_, erra := io.Copy(file, resp.Body)
	//fmt.Println(size)
	defer file.Close()
	if erra != nil {
		Logger.Errorw("Error saving file"+link, err)
		return "", erra
	}
	changeOwnership(finalPath)
	return finalPath, nil

}
func changeOwnership(path string) {
	uid, err1 := strconv.Atoi(os.Getenv("PUID"))
	gid, err2 := strconv.Atoi(os.Getenv("PGID"))
	fmt.Println(path)
	if err1 == nil && err2 == nil {
		fmt.Println(path + " : Attempting change")
		os.Chown(path, uid, gid)
	}

}
func DeleteFile(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return err
	}
	if err := os.Remove(filePath); err != nil {
		return err
	}
	return nil
}
func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil

}

func GetAllBackupFiles() ([]string, error) {
	var files []string
	folder := createIfFoldeDoesntExist("backups")
	err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	sort.Sort(sort.Reverse(sort.StringSlice(files)))
	return files, err
}

func deleteOldBackup() {
	files, err := GetAllBackupFiles()
	if err != nil {
		return
	}
	if len(files) <= 5 {
		return
	}

	toDelete := files[5:]
	for _, file := range toDelete {
		fmt.Println(file)
		DeleteFile(file)
	}
}

func CreateBackup() (string, error) {

	backupFileName := "podgrab_backup_" + time.Now().Format("2006.01.02_150405") + ".tar.gz"
	folder := createIfFoldeDoesntExist("backups")
	configPath := os.Getenv("CONFIG")
	tarballFilePath := path.Join(folder, backupFileName)
	file, err := os.Create(tarballFilePath)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Could not create tarball file '%s', got error '%s'", tarballFilePath, err.Error()))
	}
	defer file.Close()

	dbPath := path.Join(configPath, "podgrab.db")
	_, err = os.Stat(dbPath)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Could not find db file '%s', got error '%s'", dbPath, err.Error()))
	}
	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	err = addFileToTarWriter(dbPath, tarWriter)
	if err == nil {
		deleteOldBackup()
	}
	return backupFileName, err
}

func addFileToTarWriter(filePath string, tarWriter *tar.Writer) error {
	file, err := os.Open(filePath)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not open file '%s', got error '%s'", filePath, err.Error()))
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return errors.New(fmt.Sprintf("Could not get stat for file '%s', got error '%s'", filePath, err.Error()))
	}

	header := &tar.Header{
		Name:    filePath,
		Size:    stat.Size(),
		Mode:    int64(stat.Mode()),
		ModTime: stat.ModTime(),
	}

	err = tarWriter.WriteHeader(header)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not write header for file '%s', got error '%s'", filePath, err.Error()))
	}

	_, err = io.Copy(tarWriter, file)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not copy the file '%s' data to the tarball, got error '%s'", filePath, err.Error()))
	}

	return nil
}
func httpClient() *http.Client {
	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			//	r.URL.Opaque = r.URL.Path
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
		changeOwnership(folderPath)
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
