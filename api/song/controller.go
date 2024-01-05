package song

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"top_100_billboard_golang/api/common"
	"top_100_billboard_golang/repository/database"

	"github.com/gin-gonic/gin"
)

func SetupRouter(r *gin.Engine) {

	songsEndpoint := r.Group("/songs")
	songsEndpoint.Use(common.AccessTokenMiddleware)
	songsEndpoint.GET("", getSongs)
}

func getSongs(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		common.WriteResponse(c, nil, "failed to covert 'page' from string to int", http.StatusBadRequest)
		return
	}

	reversedAsString := c.DefaultQuery("reversed", "true")
	reversed := false
	if strings.ToLower(reversedAsString) == "true" {
		reversed = true
	}

	songsInfo, err := database.SongInfoWrapper.GetSongInfoList(reversed, page)
	if err != nil {
		common.WriteResponse(c, nil, fmt.Sprintf("failed query db, err: %s", err.Error()), http.StatusBadRequest)
		return
	}

	common.WriteResponse(c, songsInfo, "Success", http.StatusOK)
}
