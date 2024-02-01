package ads

import (
	"os"

	"github.com/gin-gonic/gin"
)

func SetupRouter(r *gin.Engine) {
	r.GET("/app-ads.txt", getFile)
}

func getFile(c *gin.Context) {
	c.Header("Content-Disposition", "attachment; filename=app-ads.txt")
	c.File(os.Getenv("APP_ADS_PATH"))
}
