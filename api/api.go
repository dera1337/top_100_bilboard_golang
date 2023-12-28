package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"top_100_billboard_golang/repository/database"

	"github.com/gin-gonic/gin"
)

func Run() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	// bikin tabel user (user_id, username, fcm_token 'string') //buat notifikasi
	// bikin endpoint sign up (
	// 	req: username(body),
	// 	resp: (access_token, refresh_token)(body)
	// )
	// bikin endpoint refresh access_token (
	// 	req: refresh_token (Authorization header),
	// 	resp: access_token(body)
	// )
	r.Use(AccessTokenMiddleware)
	r.GET("/", func(c *gin.Context) {
		writeResponse(c, nil, "ok", http.StatusOK)
	})
	songsEndpoint := r.Group("/songs")
	songsEndpoint.GET("", func(c *gin.Context) {
		page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
		if err != nil {
			writeResponse(c, nil, "failed to covert 'page' from string to int", http.StatusBadRequest)
			return
		}

		reversedAsString := c.DefaultQuery("reversed", "true")
		reversed := false
		if strings.ToLower(reversedAsString) == "true" {
			reversed = true
		}

		songsInfo, err := database.SongInfoWrapper.GetSongInfoList(reversed, page)
		if err != nil {
			writeResponse(c, nil, fmt.Sprintf("failed query db, err: %s", err.Error()), http.StatusBadRequest)
			return
		}

		writeResponse(c, songsInfo, "Success", http.StatusOK)
	})
	r.Run("localhost:8080")
}

type genericResponse struct {
	Data       interface{} `json:"data"`
	Message    string      `json:"message"`
	StatusCode int         `json:"status_code"`
}

func writeResponse(
	c *gin.Context,
	data interface{},
	msg string,
	statusCode int,
) {
	resp := genericResponse{
		Data:       data,
		Message:    msg,
		StatusCode: statusCode,
	}

	c.JSON(statusCode, resp)
}
