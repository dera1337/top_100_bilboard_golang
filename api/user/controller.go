package user

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"top_100_billboard_golang/api/common"
	"top_100_billboard_golang/repository/database"

	"github.com/gin-gonic/gin"
)

func SetupRouter(r *gin.Engine) {

	usersEndpoint := r.Group("/users")
	usersEndpoint.POST("/signup", signUp)
	usersEndpoint.POST("/refresh", refreshToken)
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
		common.WriteResponse(c, nil, "failed to insert user to database", http.StatusBadRequest)
		return
	}

	accessToken, err := common.GenerateAccess(req.IsPremium)
	if err != nil {
		common.WriteResponse(c, nil, "failed to generate token", http.StatusInternalServerError)
		return
	}
	refreshToken, err := common.GenerateRefresh(req.IsPremium)
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

func refreshToken(c *gin.Context) {
	jsonBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		common.WriteResponse(c, nil, "failed reading request body", http.StatusBadRequest)
		return
	}
	defer c.Request.Body.Close()

	var req database.RefreshTokenRequest
	err = json.Unmarshal(jsonBytes, &req)
	if err != nil {
		common.WriteResponse(c, nil, "failed reading request body", http.StatusBadRequest)
		return
	}

	claims, err := common.ParseJWT(c)
	if err != nil {
		common.WriteResponse(c, nil, err.Error(), http.StatusUnauthorized)
		return
	}

	tokenType, err := strconv.Atoi(fmt.Sprintf("%v", claims["token_type"]))
	if err != nil {
		common.WriteResponse(c, nil, err.Error(), http.StatusUnauthorized)
		return
	}
	if common.TokenType(tokenType) != common.Refresh {
		common.WriteResponse(c, nil, "Wrong token type", http.StatusUnauthorized)
		return
	}

	isPremium, ok := claims["is_premium"].(bool)
	if !ok {
		common.WriteResponse(c, nil, "Invalid claim found in token", http.StatusUnauthorized)
		return
	}

	accessToken, err := common.GenerateAccess(isPremium)
	if err != nil {
		common.WriteResponse(c, nil, "Failed to generate access token", http.StatusBadRequest)
		return
	}

	resp := database.RefreshTokenResponse{
		AccessToken: accessToken,
	}

	common.WriteResponse(c, &resp, "Access Token is Generated", http.StatusOK)

}
