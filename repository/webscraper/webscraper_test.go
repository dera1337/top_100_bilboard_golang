package webscraper

import (
	"testing"
	"top_100_billboard_golang/repository/database"

	"github.com/stretchr/testify/assert"
)

func TestCacheSongs(t *testing.T) {
	datas := []database.SongInformation{
		{Id: 1}, {Id: 2}, {Id: 3},
	}
	cacheSongs(datas)

	// odd
	for i := 0; i < len(datas); i++ {
		assert.Equal(t, datas[i].Id, CachedSongTitles[i].Id)
		assert.Equal(t, datas[len(datas)-i-1].Id, CachedSongTitlesReversed[i].Id)
	}

	// even
	datas = append(datas, database.SongInformation{Id: 4})
	cacheSongs(datas)
	for i := 0; i < len(datas); i++ {
		assert.Equal(t, datas[i].Id, CachedSongTitles[i].Id)
		assert.Equal(t, datas[len(datas)-i-1].Id, CachedSongTitlesReversed[i].Id)
	}
}
