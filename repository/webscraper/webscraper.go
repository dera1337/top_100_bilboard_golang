package webscraper

import (
	"log"
	"net/http"
	"os"
	"strings"
	"top_100_billboard_golang/repository/database"

	"github.com/PuerkitoBio/goquery"
)

var (
	errCount         int = 1
	CachedSongTitles []string
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
		imageUrl, _ := s.Attr("style")
		index := strings.Index(imageUrl, "https")
		lastIndex := strings.LastIndex(imageUrl, "'")
		imageUrl = imageUrl[index:lastIndex]
		songInfo := database.SongInformation{
			ImageURL: imageUrl,
		}
		songInfoList = append(songInfoList, songInfo)
	})

	doc.Find(os.Getenv("SELECTOR_2")).Each(func(i int, s *goquery.Selection) {
		title := strings.TrimSpace(s.Find("h3").Text())
		author := strings.TrimSpace(s.Find("span.c-label").Text())

		songInfo := songInfoList[i]
		songInfo.Title = title
		songInfo.Artist = author
		songInfo.CurrentRank = i + 1

		songInfoList[i] = songInfo

	})

	// populate variable CachedSongTitles dari db
	if len(CachedSongTitles) == 0 {
		CachedSongTitles = database.SongInfoWrapper.GetCacheFromDB()
	}

	// apakah di db ada isi / no, kalo ngga
	if len(CachedSongTitles) == 0 {
		// print "no data inside db"
		// fmt.Println("no data inside db")
		for _, v := range songInfoList {
			err := database.SongInfoWrapper.GetVideoList(&v)
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
			if songInfoList[i].Title != CachedSongTitles[i] {

				// print "data difference detected"
				// fmt.Println("data difference detected")
				for _, v := range songInfoList {
					err := database.SongInfoWrapper.GetVideoList(&v)
					if err != nil {
						log.Fatal(err)
					}
				}

				err = database.SongInfoWrapper.InsertRows(songInfoList)
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

// for cache 25 song so we don't have to querry db
func cacheSongs(songs []database.SongInformation) {
	result := []string{}

	for _, v := range songs[:25] {

		result = append(result, v.Title)

	}
	CachedSongTitles = result
}
