package song

import (
	"testing"
	"top_100_billboard_golang/repository/database"
	"top_100_billboard_golang/repository/webscraper"

	"github.com/stretchr/testify/assert"
)

func TestPaginateSongInfoList(t *testing.T) {
	webscraper.CachedSongTitles = make([]database.SongInformation, 100)
	webscraper.CachedSongTitlesReversed = make([]database.SongInformation, 100)
	for i := 0; i < 100; i++ {
		webscraper.CachedSongTitles[i] = database.SongInformation{Id: i + 1}
		webscraper.CachedSongTitlesReversed[100-i-1] = database.SongInformation{Id: i + 1}
	}

	result := paginateSongInfoList(1, true)
	assert.Equal(t, 25, len(result))
	assert.Equal(t, 1, result[0].Id)
	assert.Equal(t, 25, result[24].Id)

	result = paginateSongInfoList(2, true)
	assert.Equal(t, 25, len(result))
	assert.Equal(t, 26, result[0].Id)
	assert.Equal(t, 50, result[24].Id)

	result = paginateSongInfoList(3, true)
	assert.Equal(t, 25, len(result))
	assert.Equal(t, 51, result[0].Id)
	assert.Equal(t, 75, result[24].Id)

	result = paginateSongInfoList(4, true)
	assert.Equal(t, 25, len(result))
	assert.Equal(t, 76, result[0].Id)
	assert.Equal(t, 100, result[24].Id)

	result = paginateSongInfoList(5, true)
	assert.Equal(t, 0, len(result))

	result = paginateSongInfoList(1, false)
	assert.Equal(t, 25, len(result))
	assert.Equal(t, 100, result[0].Id)
	assert.Equal(t, 76, result[24].Id)

	result = paginateSongInfoList(2, false)
	assert.Equal(t, 25, len(result))
	assert.Equal(t, 75, result[0].Id)
	assert.Equal(t, 51, result[24].Id)

	result = paginateSongInfoList(3, false)
	assert.Equal(t, 25, len(result))
	assert.Equal(t, 50, result[0].Id)
	assert.Equal(t, 26, result[24].Id)

	result = paginateSongInfoList(4, false)
	assert.Equal(t, 25, len(result))
	assert.Equal(t, 25, result[0].Id)
	assert.Equal(t, 1, result[24].Id)

	result = paginateSongInfoList(5, false)
	assert.Equal(t, 0, len(result))
}
