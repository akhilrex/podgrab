package db

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type localMigration struct {
	Name  	  string
	Condition []string
	Query	  []string
}

var migrations = []localMigration{
	{
		Name:  	   "2020_11_03_04_42_SetDefaultDownloadStatus",
		Condition: []string{},
		Query:     []string{
			"update podcast_items set download_status=2 where download_path!='' and download_status=0",
		},
	},
	{
		Name:  "2021_06_01_00_00_ConvertFileNameFormat",
		Condition: []string{
			"SELECT COUNT(*) > 0 FROM (SELECT name FROM pragma_table_info('settings') where name is 'append_date_to_file_name');",
			"SELECT COUNT(*) > 0 FROM (SELECT name FROM pragma_table_info('settings') where name is 'append_episode_number_to_file_name');",
		},
		Query: []string{
			"UPDATE settings SET file_name_format = CASE WHEN append_date_to_file_name AND append_episode_number_to_file_name THEN '%EpisodeNumber%-%EpisodeDate%-%EpisodeTitle%' WHEN append_date_to_file_name THEN '%EpisodeDate%-%EpisodeTitle%' WHEN append_episode_number_to_file_name THEN '%EpisodeNumber%-%EpisodeTitle%' ELSE '%EpisodeTitle%' END",
// sqlite3 v3.35.0 supports DROP COLUMN
// 			"ALTER TABLE settings DROP COLUMN append_episode_number_to_file_name",
// 			"ALTER TABLE settings DROP COLUMN append_date_to_file_name",
		},
	},
}

func RunMigrations() {
	for _, mig := range migrations {
		fmt.Println(mig.Name)
		ExecuteAndSaveMigration(mig.Name, mig.Condition, mig.Query)
	}
}
func ExecuteAndSaveMigration(name string, condition []string, query []string) error {
	var migration Migration
	result := DB.Where("name=?", name).First(&migration)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		var rawResult string
		var shouldMigrate = true
		for _, q := range condition {
			fmt.Println("cond: " + q)
			result = DB.Debug().Raw(q).Scan(&rawResult)
			if result.Error != nil {
				fmt.Println(result.Error)
				return result.Error
			}
			shouldMigrate = shouldMigrate && rawResult == "1"
		}
		if shouldMigrate {
			for _, q := range query {
				fmt.Println("exec: " + q)
				result = DB.Debug().Exec(q)
				if result.Error != nil {
					fmt.Println(result.Error)
					return result.Error
				}
			}
		} else {
			fmt.Println("migration not required")
		}
		DB.Save(&Migration{
			Date: time.Now(),
			Name: name,
		})
		return result.Error
	} else {
		fmt.Println("skipping")
	}
	return nil
}
