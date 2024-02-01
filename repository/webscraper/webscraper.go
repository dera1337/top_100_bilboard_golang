package webscraper

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"top_100_billboard_golang/notification"
	"top_100_billboard_golang/repository/database"

	"github.com/PuerkitoBio/goquery"
)

var (
	errCount                 int = 1
	CachedSongTitles         []database.SongInformation
	CachedSongTitlesReversed []database.SongInformation
)

func ScrapeBillboard() {
	log.Println("scraping begin")

	var res *http.Response
	for {
		result, err := http.Get(os.Getenv("URLBILLBOARD"))
		if err != nil {
			if errCount <= 5 {
				// print "increment error count"
				// fmt.Println("increment error count")
				errCount++
				continue
			} else {
				// print "failed to request billboard site"
				// fmt.Println("failed to request billboard site")
				log.Println(err)
				return
			}
		}
		// print "positive flow"
		// fmt.Println("positive flow")

		res = result
		errCount = 1
		break
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	songInfoList := []database.SongInformation{}
	doc.Find(os.Getenv("SELECTOR_1")).Each(func(i int, s *goquery.Selection) {
		imageUrl, ok := s.Attr("data-lazy-src")
		if !ok {
			log.Fatal(fmt.Errorf("image scraper selector broken"))
		}
		songInfo := database.SongInformation{
			ImageURL: imageUrl,
		}
		songInfoList = append(songInfoList, songInfo)
	})

	doc.Find(os.Getenv("SELECTOR_2")).Each(func(i int, s *goquery.Selection) {
		title := strings.TrimSpace(s.Find("h3").Text())
		author := strings.TrimSpace(s.Find("span.c-label").Text())

		songInfoList[i].Title = title
		songInfoList[i].Artist = author
		songInfoList[i].CurrentRank = i + 1
	})

	// apakah di db ada isi / no, kalo ngga
	if len(CachedSongTitles) == 0 {
		// print "no data inside db"
		// fmt.Println("no data inside db")
		for i := 0; i < len(songInfoList); i++ {
			err := database.SongInfoWrapper.GetVideoList(&songInfoList[i])
			if err != nil {
				log.Fatal(err)
			}
		}

		err = database.SongInfoWrapper.InsertRows(songInfoList)
		if err != nil {
			log.Fatal(err)
		}
		cacheSongs(songInfoList)
	} else {
		// print "cache exist in db"
		// fmt.Println("cache exist in db")
		for i := 0; i < len(CachedSongTitles); i++ {
			if songInfoList[i].Title != CachedSongTitles[i].Title {

				// print "data difference detected"
				// fmt.Println("data difference detected")
				for j := 0; j < len(songInfoList); j++ {
					err := database.SongInfoWrapper.GetVideoList(&songInfoList[j])
					if err != nil {
						log.Fatal(err)
					}
				}

				err = database.SongInfoWrapper.InsertRows(songInfoList)
				if err != nil {
					log.Fatal(err)
				}

				err = notification.SendNotificationMessageToPaidUsers(
					songInfoList[0].Artist,
				)
				if err != nil {
					log.Fatal(err)
				}

				// fmt.Println("after inserting rows")
				cacheSongs(songInfoList)
				return
			}
		}
	}
}

// populate variable CachedSongTitles dari db
func PopulateCache() {
	cacheSongs(database.SongInfoWrapper.GetCacheFromDB())
}

// for cache 100 songs so we don't have to query db
func cacheSongs(songs []database.SongInformation) {
	length := len(songs)

	CachedSongTitles = songs
	CachedSongTitlesReversed = make([]database.SongInformation, length)

	j := 0
	for i := length - 1; i >= 0; i-- {
		CachedSongTitlesReversed[j] = songs[i]
		j++
	}
}
