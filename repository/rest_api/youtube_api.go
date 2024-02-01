package restapi

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
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

func (vd *VideoDetail) GetYoutubeUrl(title, artist string) (string, error) {
	artist = strings.ToLower(artist)
	idx := strings.Index(artist, ",")
	commaExist := idx != -1
	if commaExist {
		artist = artist[:idx]
	} else {
		artistSplit := strings.Split(artist, " ")
	loop:
		for i, v := range artistSplit {
			switch v {
			case "featuring", "&", "with", "x":
				artist = strings.Join(artistSplit[:i], " ")
				break loop
			}
		}
	}

	cache := make(map[int]int)
	order := []int{}

	for i := 0; i < len(vd.Items); i++ {
		channelNameMatch := true

		channelName := strings.ReplaceAll(strings.ToLower(vd.Items[i].Snippet.ChannelTitle), " ", "")
		for _, v := range strings.Split(artist, " ") {
			if !strings.Contains(channelName, v) {
				channelNameMatch = false
			}
		}

		if channelNameMatch {
			jsonBytes := GetVideoStatistic(vd.Items[i].Id.VideoId, ApiKey1)
			if jsonBytes == nil {
				jsonBytes = GetVideoStatistic(vd.Items[i].Id.VideoId, ApiKey2)

				if jsonBytes == nil {
					return "", fmt.Errorf("both keys reached quotas")
				}
			}

			videoStat := videoStatistic{}

			err := json.Unmarshal(jsonBytes, &videoStat)
			if err != nil {
				return "", err
			}

			viewCount, err := strconv.Atoi(videoStat.Items[0].Statistics.ViewCount)
			if err != nil {
				return "", err
			}
			vd.Items[i].ViewCount = viewCount

			processedTitle := strings.ToLower(vd.Items[i].Snippet.Title)
			for _, v := range strings.Split(artist, " ") {
				processedTitle = strings.Replace(processedTitle, v, "", 1)
			}

			titleMatchPercentage := countSubsetString(processedTitle, title)
			vd.Items[i].MatchPercentage = titleMatchPercentage

			if viewCount <= 750000 {
				continue
			}

			idx, ok := cache[titleMatchPercentage]
			if ok {
				if viewCount > vd.Items[idx].ViewCount {
					cache[titleMatchPercentage] = i
				}
			}
			cache[titleMatchPercentage] = i
			order = append(order, titleMatchPercentage)
		}
	}

	if len(order) == 0 {
		return fmt.Sprintf("https://www.youtube.com/watch?v=%s", vd.Items[0].Id.VideoId), nil
	}

	sort.Ints(order)

	return fmt.Sprintf("https://www.youtube.com/watch?v=%s", vd.Items[cache[len(order)-1]].Id.VideoId), nil
}

type item struct {
	Id              id      `json:"id"`
	Snippet         snippet `json:"snippet"`
	ViewCount       int
	MatchPercentage int
}

type id struct {
	VideoId string `json:"videoId"`
}

type snippet struct {
	Title        string `json:"title"`
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
	url := createEncodedURL(
		os.Getenv("GOOGLE_API_LINK_1"),
		fmt.Sprintf("%s %s", author, title),
		apiKey,
	)

	res, err := http.Get(url)
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

func createEncodedURL(baseURL, searchQuery, apiKey string) string {
	queryParams := map[string]string{
		"q":          searchQuery,
		"key":        apiKey,
		"type":       "video",
		"part":       "id,snippet",
		"maxResults": "5",
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		log.Fatal(err)
	}

	q := u.Query()
	for k, v := range queryParams {
		q.Add(k, v)
	}

	u.RawQuery = q.Encode()
	return u.String()
}

func countSubsetString(string1, string2 string) int {
	string1 = strings.ReplaceAll(strings.ToLower(string1), " ", "")
	string2 = strings.ReplaceAll(strings.ToLower(string2), " ", "")

	l := len(string2) + 1
	ll := len(string1) + 1

	maxVal := 0

	matrix := make([][]int, l)
	for i := 0; i < l; i++ {
		matrix[i] = make([]int, ll)
	}

	for i := 1; i < l; i++ {
		for j := 1; j < ll; j++ {
			if string2[i-1] == string1[j-1] {
				val := matrix[i-1][j-1] + 1
				matrix[i][j] = val
				if maxVal < val {
					maxVal = val
				}
			} else {
				matrix[i][j] = max(matrix[i-1][j], matrix[i][j-1])
			}
		}
	}

	return maxVal*100/min(l, ll) - 1
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
