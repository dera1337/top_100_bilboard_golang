package restapi

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var (
	ApiKey1 string
	ApiKey2 string
)

func PopulateApiKey() {
	ApiKey1 = os.Getenv("APIKEY_1")
	ApiKey2 = os.Getenv("APIKEY_2")
}

type VideoDetail struct {
	Items []item `json:"items"`
}

func (vd *VideoDetail) GetYoutubeUrl() (string, error) {
	var videoStatPlaceholder videoStatistic

	for _, value := range vd.Items {
		jsonBytes := GetVideoStatistic(value.Id.VideoId, ApiKey1)
		if jsonBytes == nil {
			jsonBytes = GetVideoStatistic(value.Id.VideoId, ApiKey2)

			if jsonBytes == nil {
				return "", fmt.Errorf("both keys reached quotas")
			}
		}

		videoStat := videoStatistic{}

		err := json.Unmarshal(jsonBytes, &videoStat)

		if err != nil {
			return "", err
		}
		if len(videoStatPlaceholder.Items) == 0 {
			videoStatPlaceholder = videoStat
		} else {
			currentView, _ := strconv.Atoi(videoStat.Items[0].Statistics.ViewCount)
			placeholderView, _ := strconv.Atoi(videoStatPlaceholder.Items[0].Statistics.ViewCount)

			if currentView > placeholderView {
				videoStatPlaceholder = videoStat
			}
		}

	}
	return fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoStatPlaceholder.Items[0].Id), nil
}

type item struct {
	Id      id      `json:"id"`
	Snippet snippet `json:"snippet"`
}

type id struct {
	VideoId string `json:"videoId"`
}

type snippet struct {
	ChannelTitle string `json:"channelTitle"`
}

type videoStatistic struct {
	Items []itemStat `json:"items"`
}

type itemStat struct {
	Id         string `json:"id"`
	Statistics stat   `json:"statistics"`
}

type stat struct {
	ViewCount string `json:"viewCount"`
}

// this function to get videoId
func GetVideoDetail(title, author, apiKey string) []byte {

	url := os.Getenv("GOOGLE_API_LINK_1")

	finalUrl := fmt.Sprintf(url, author+" "+title, apiKey)
	finalUrl = strings.Replace(finalUrl, " ", "%20", -1)

	// fmt.Println(finalUrl)

	res, err := http.Get(finalUrl)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		if res.StatusCode != 403 {
			log.Printf(
				"another error occured that is not caused by API key quota, status code: %d %s\n",
				res.StatusCode,
				res.Status,
			)
		}
		return nil
	}

	jsonBytes, err := io.ReadAll(res.Body)

	if err != nil {
		log.Println(err)
		return nil
	}
	return jsonBytes
}

// this is function to get viewCount
func GetVideoStatistic(id, key string) []byte {
	url := os.Getenv("GOOGLE_API_LINK_2")

	finalUrl := fmt.Sprintf(url, id, key)
	// fmt.Println(finalUrl)

	res, err := http.Get(finalUrl)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		if res.StatusCode != 403 {
			log.Printf(
				"another error occured that is not caused by API key quota, status code: %d %s\n",
				res.StatusCode,
				res.Status,
			)
		}
		return nil
	}

	jsonBytes, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		return nil
	}
	return jsonBytes
}
