package user

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"top_100_billboard_golang/api/common"
	"top_100_billboard_golang/repository/database"

	"github.com/gin-gonic/gin"
)

func SetupRouter(r *gin.Engine) {

	usersEndpoint := r.Group("/users")
	usersEndpoint.POST("/signup", signUp)
}

func signUp(c *gin.Context) {
	jsonBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		common.WriteResponse(c, nil, "failed reading request body", http.StatusBadRequest)
		return
	}
	defer c.Request.Body.Close()

	var req database.SignUpRequest
	err = json.Unmarshal(jsonBytes, &req)
	if err != nil {
		common.WriteResponse(c, nil, "failed reading request body", http.StatusBadRequest)
		return
	}

	// TODO: send & read the response body from FCM to determine whether user
	// is using our app or pirated app

	// insert db
	err = database.UserWrapper.InsertUser(&req)
	if err != nil {
		fmt.Println(err.Error())
		common.WriteResponse(c, nil, "failed to insert user to database", http.StatusBadRequest)
		return
	}

	accessToken, err := common.GenerateAccess()
	if err != nil {
		common.WriteResponse(c, nil, "failed to generate token", http.StatusInternalServerError)
		return
	}
	refreshToken, err := common.GenerateRefresh()
	if err != nil {
		common.WriteResponse(c, nil, "failed to generate token", http.StatusInternalServerError)
		return
	}

	resp := database.SignUpResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	common.WriteResponse(c, &resp, "success sign up", http.StatusOK)
}
