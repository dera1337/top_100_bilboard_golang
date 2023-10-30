package database

import (
	"encoding/json"
	"fmt"
	"log"
	restapi "top_100_billboard_golang/repository/rest_api"

	"github.com/jackc/pgx/v5"
)

type SongInformation struct {
	Title        string
	Artist       string
	Link         string
	ImageURL     string
	CurrentRank  int
	PreviousRank int
}

type songInfoWrapper struct {
	dbConn
}

var (
	SongInfoWrapper songInfoWrapper
)

func (siw *songInfoWrapper) GetVideoList(si *SongInformation) error {

	jsonBytes := restapi.GetVideoDetail(si.Title, si.Artist, restapi.ApiKey1)
	if jsonBytes == nil {
		jsonBytes = restapi.GetVideoDetail(si.Title, si.Artist, restapi.ApiKey2)

		if jsonBytes == nil {
			return fmt.Errorf("both keys reached quotas")
		}
	}

	videoDetail := restapi.VideoDetail{}

	err := json.Unmarshal(jsonBytes, &videoDetail)
	if err != nil {
		return err
	}

	url, err := videoDetail.GetYoutubeUrl()
	if err != nil {
		return err
	}

	si.Link = url
	return nil
}

func (siw *songInfoWrapper) GetCacheFromDB() []string {
	rowCount := siw.getTableCount()

	if rowCount <= 0 {
		return nil
	}

	rows, err := conn.Pool.Query(conn.Ctx, "select title * FROM song_information ORDER BY rank_number asc")

	if err != nil {
		log.Println(err)
		return nil
	}

	result := []string{}

	for rows.Next() {
		var title string
		err := rows.Scan(&title)
		if err != nil {
			log.Println(err)
			return nil
		}
		result = append(result, title)
	}
	return result
}

func (siw *songInfoWrapper) InsertRows(songInfoList []SongInformation) error {
	songInfoList, err := siw.updatePreviousRank(songInfoList)
	if err != nil {
		return err
	}

	rows := convertToInterface(songInfoList)

	copyCount, err := siw.Pool.CopyFrom(
		conn.Ctx,
		pgx.Identifier{"song_information"},
		[]string{"title", "author", "youtube_url", "image_url", "current_rank_number", "previous_rank_number"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return err
	}

	if copyCount != 100 {
		return fmt.Errorf("expected 100 but got: %d", copyCount)
	}

	// insertStatement := "INSERT INTO song_information (title, author, youtube_url, image_url) VALUES ($1, $2, $3, $4)"

	// // conn.Pool.Exec(conn.Ctx, "INSERT INTO song_information (title, author, youtube_url) VALUES (abcdef, wdas, http:www.)")
	// conn.Pool.Exec(conn.Ctx, insertStatement, song.Title, song.Artist, song.Link, song.ImageURL)
	return nil
}

func (siw *songInfoWrapper) updatePreviousRank(songInfoList []SongInformation) ([]SongInformation, error) {
	rowCount := siw.getTableCount()
	if rowCount <= 0 {
		return songInfoList, nil
	}

	rows, err := conn.Pool.Query(conn.Ctx, "select * FROM song_information ORDER BY current_rank_number asc")
	if err != nil {
		log.Println(err)
		return nil, err
	}

	dbSongInfoList := []SongInformation{}

	for rows.Next() {
		var id int
		var songInfo SongInformation
		err := rows.Scan(
			&id,
			&songInfo.Title,
			&songInfo.Artist,
			&songInfo.Link,
			&songInfo.ImageURL,
			&songInfo.CurrentRank,
			&songInfo.PreviousRank,
		)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		dbSongInfoList = append(dbSongInfoList, songInfo)
	}

loop:
	for i, v := range songInfoList {
		for _, vv := range dbSongInfoList {
			if v.Title == vv.Title {
				songInfoList[i].PreviousRank = vv.CurrentRank
				continue loop
			}
		}
	}
	return songInfoList, nil
}

func (siw *songInfoWrapper) getTableCount() int {
	rowCount := 0
	row := siw.Pool.QueryRow(conn.Ctx, "select count(*) FROM song_information")
	err := row.Scan(&rowCount)
	if err != nil {
		log.Println(err)
		return 0
	}
	return rowCount
}

func convertToInterface(songInfoList []SongInformation) [][]interface{} {
	result := [][]interface{}{}

	for idx, v := range songInfoList {
		innerList := []interface{}{
			v.Title,
			v.Artist,
			v.Link,
			v.ImageURL,
			idx + 1,
		}

		if v.PreviousRank != 0 {
			innerList = append(innerList, v.PreviousRank)
		} else {
			innerList = append(innerList, nil)
		}

		result = append(result, innerList)
	}
	return result

}
