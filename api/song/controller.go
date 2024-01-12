package song

import (
	"fmt"
	"net/http"
	"strconv"
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

	claims, err := common.ParseJWT(c)
	if err != nil {
		common.WriteResponse(c, nil, err.Error(), http.StatusUnauthorized)
		return
	}

	isPremium, ok := claims["is_premium"].(bool)
	if !ok {
		common.WriteResponse(c, nil, "Invalid claim found in token", http.StatusUnauthorized)
		return
	}

	reversed := !isPremium

	songsInfo, err := database.SongInfoWrapper.GetSongInfoList(reversed, page)
	if err != nil {
		common.WriteResponse(c, nil, fmt.Sprintf("failed query db, err: %s", err.Error()), http.StatusBadRequest)
		return
	}

	common.WriteResponse(c, songsInfo, "Success", http.StatusOK)
}
