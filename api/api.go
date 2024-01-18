package api

import (
	"fmt"
	"os"
	"top_100_billboard_golang/api/song"
	"top_100_billboard_golang/api/user"

	"github.com/gin-gonic/gin"
)

func Run() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	setupRouter(r)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(fmt.Sprintf(":%s", port))
}

func setupRouter(r *gin.Engine) {
	user.SetupRouter(r)
	song.SetupRouter(r)
}
