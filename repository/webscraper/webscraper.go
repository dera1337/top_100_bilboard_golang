package webscraper

import (
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"top_100_billboard_golang/repository/database"

	"github.com/PuerkitoBio/goquery"
)

var (
	errCount         int32 = 1
	CachedSongTitles []string
)

func ScrapeBillboard() {

	res, err := http.Get(os.Getenv("URLBILLBOARD")) //
	if err != nil {
		if errCount <= 5 {
			atomic.AddInt32(&errCount, 1)
		} else {
			log.Println(err)
			return
		}
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	atomic.StoreInt32(&errCount, 1)

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	songInfoList := []database.SongInformation{}
	doc.Find(os.Getenv("SELECTOR_1")).Each(func(i int, s *goquery.Selection) { //
		imageUrl, _ := s.Attr("style")
		index := strings.Index(imageUrl, "https")
		lastIndex := strings.LastIndex(imageUrl, "'")
		imageUrl = imageUrl[index:lastIndex]
		songInfo := database.SongInformation{
			ImageURL: imageUrl,
		}
		songInfoList = append(songInfoList, songInfo)
	})

	i := 0
	doc.Find(os.Getenv("SELECTOR_2")).Each(func(i int, s *goquery.Selection) { //
		title := strings.TrimSpace(s.Find("h3").Text())
		author := strings.TrimSpace(s.Find("span.c-label").Text())

		songInfo := songInfoList[i]
		songInfo.Title = title
		songInfo.Artist = author

		err := database.SongInfoWrapper.GetVideoList(&songInfo)
		if err != nil {
			log.Fatal(err)
		}

		songInfoList[i] = songInfo

		i++
	})

	if len(CachedSongTitles) == 0 {
		CachedSongTitles = database.SongInfoWrapper.GetCacheFromDB()
	}

	if len(CachedSongTitles) == 0 {
		err := database.SongInfoWrapper.InsertRows(songInfoList)
		if err != nil {
			log.Fatal(err)
		}
		cacheSongs(songInfoList)
	} else {

		// kalo != 0 -> cek sama cache apakah ada perbedaan?
		// Yes = insert + populate cache, No = skip
	}

	doNothing(i)
}

// for cache 25 song so we don't have to querry db
func cacheSongs(songs []database.SongInformation) {
	result := []string{}

	for _, v := range songs[:25] {

		result = append(result, v.Title)

	}
	CachedSongTitles = result
}

func doNothing(i int) {

}
