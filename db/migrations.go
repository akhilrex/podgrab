package db

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type localMigration struct {
	Name  string
	Query string
}

var migrations = []localMigration{
	{
		Name:  "2020_11_03_04_42_SetDefaultDownloadStatus",
		Query: "update podcast_items set download_status=2 where download_path!='' and download_status=0",
	},
	{
		Name:  "2025_09_17_14_00_AddPassthroughPodcastGuidSetting",
		Query: "ALTER TABLE settings ADD COLUMN passthroughPodcastGuid BOOLEAN NOT NULL DEFAULT FALSE; ",
	},
}

func RunMigrations() {
	for _, mig := range migrations {
		ExecuteAndSaveMigration(mig.Name, mig.Query)
	}
}
func ExecuteAndSaveMigration(name string, query string) error {
	var migration Migration
	result := DB.Where("name=?", name).First(&migration)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		fmt.Println(query)
		result = DB.Debug().Exec(query)
		if result.Error == nil {
			DB.Save(&Migration{
				Date: time.Now(),
				Name: name,
			})
		}
		return result.Error
	}
	return nil
}
