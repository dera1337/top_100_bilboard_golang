package database

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	restapi "top_100_billboard_golang/repository/rest_api"

	"github.com/jackc/pgx/v5"
)

type SongInformation struct {
	Id           int    `json:"id"`
	Title        string `json:"title"`
	Artist       string `json:"artist"`
	Link         string `json:"link"`
	ImageURL     string `json:"image_url"`
	CurrentRank  int    `json:"current_rank"`
	PreviousRank *int   `json:"previous_rank"`
}

type songInfoWrapper struct {
	*dbConn
	sync.RWMutex
}

var (
	SongInfoWrapper songInfoWrapper
)

func (siw *songInfoWrapper) GetSongInfoList(reversed bool, page int) ([]SongInformation, error) {
	offset := (page - 1) * 25

	var sql string
	if reversed {
		sql = fmt.Sprintf("select * FROM song_information ORDER BY current_rank_number DESC LIMIT 25 OFFSET %d", offset)
	} else {
		sql = fmt.Sprintf("select * FROM song_information ORDER BY current_rank_number ASC LIMIT 25 OFFSET %d", offset)
	}

	siw.RLock()
	defer siw.RUnlock()

	rowsCount := siw.getTableCount()
	if rowsCount == 0 {
		return nil, nil
	}

	rows, err := siw.Pool.Query(siw.Ctx, sql)
	if err != nil {
		return nil, err
	}

	dbSongInfoList := []SongInformation{}

	for rows.Next() {
		var songInfo SongInformation
		err := rows.Scan(
			&songInfo.Id,
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
	return dbSongInfoList, nil
}

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

	url, err := videoDetail.GetYoutubeUrl(si.Title, si.Artist)
	if err != nil {
		return err
	}

	si.Link = url
	return nil
}

func (siw *songInfoWrapper) GetCacheFromDB() []SongInformation {
	rowCount := siw.getTableCount()

	if rowCount <= 0 {
		return nil
	}

	rows, err := siw.Pool.Query(
		siw.Ctx,
		`SELECT * FROM song_information ORDER BY current_rank_number ASC`,
	)

	if err != nil {
		log.Println(err)
		return nil
	}

	result := []SongInformation{}

	for rows.Next() {
		var songInfo SongInformation
		err := rows.Scan(
			&songInfo.Id,
			&songInfo.Title,
			&songInfo.Artist,
			&songInfo.Link,
			&songInfo.ImageURL,
			&songInfo.CurrentRank,
			&songInfo.PreviousRank,
		)
		if err != nil {
			log.Println(err)
			return nil
		}
		result = append(result, songInfo)
	}
	return result
}

func (siw *songInfoWrapper) InsertRows(songInfoList []SongInformation) error {
	// fmt.Println("inserting rows func")
	songInfoList, err := siw.updatePreviousRank(songInfoList)
	if err != nil {
		return err
	}

	rows := convertToInterface(songInfoList)

	siw.Lock()
	defer siw.Unlock()

	tx, err := siw.Pool.Begin(siw.Ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(siw.Ctx)

	err = siw.deleteRows(tx)
	if err != nil {
		return err
	}

	copyCount, err := tx.CopyFrom(
		siw.Ctx,
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
	return nil
}

func (siw *songInfoWrapper) deleteRows(tx pgx.Tx) error {
	sql := "DELETE FROM song_information"
	if tx == nil {
		_, err := siw.Pool.Exec(siw.Ctx, sql)
		if err != nil {
			return err
		}

		return nil
	}

	_, err := tx.Exec(siw.Ctx, sql)
	if err != nil {
		return err
	}

	return nil
}

func (siw *songInfoWrapper) updatePreviousRank(songInfoList []SongInformation) ([]SongInformation, error) {
	rowCount := siw.getTableCount()
	if rowCount <= 0 {
		return songInfoList, nil
	}
	// fmt.Println("finish query table count")

	rows, err := siw.Pool.Query(siw.Ctx, "select * FROM song_information ORDER BY current_rank_number asc")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	// fmt.Println("finish query")

	dbSongInfoList := []SongInformation{}

	// fmt.Println("start populating dbsonginfolist")
	for rows.Next() {
		var songInfo SongInformation
		err := rows.Scan(
			&songInfo.Id,
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
	// fmt.Println("finished populating dbsonginfolist")

loop:
	for i, v := range songInfoList {
		for _, vv := range dbSongInfoList {
			if v.Title == vv.Title {
				songInfoList[i].PreviousRank = &vv.CurrentRank
				// fmt.Println(*songInfoList[i].PreviousRank)
				continue loop
			}
		}
	}
	return songInfoList, nil
}

func (siw *songInfoWrapper) getTableCount() int {
	rowCount := 0
	row := siw.Pool.QueryRow(siw.Ctx, "select count(*) FROM song_information")
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
			v.PreviousRank,
		}

		result = append(result, innerList)
	}
	return result

}
