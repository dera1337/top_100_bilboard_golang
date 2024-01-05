package api

import (
	"top_100_billboard_golang/api/song"
	"top_100_billboard_golang/api/user"

	"github.com/gin-gonic/gin"
)

func Run() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	setupRouter(r)
	r.Run("localhost:8080")
}

func setupRouter(r *gin.Engine) {
	user.SetupRouter(r)
	song.SetupRouter(r)
}
