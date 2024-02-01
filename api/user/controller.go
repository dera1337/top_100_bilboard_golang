package user

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"top_100_billboard_golang/api/common"
	"top_100_billboard_golang/environment"
	"top_100_billboard_golang/notification"
	"top_100_billboard_golang/repository/database"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
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

	var projectID string
	isPremium := req.PackageName == os.Getenv("PAID_PKG_NAME")
	if isPremium {
		projectID = os.Getenv("PAID_PROJECT_ID")
	} else if req.PackageName == os.Getenv("FREE_PKG_NAME") {
		projectID = os.Getenv("FREE_PROJECT_ID")
	} else {
		common.WriteResponse(c, nil, "invalid request", http.StatusBadRequest)
		return
	}

	_, err = notification.SendNotification(
		notification.Register,
		req.FCMToken,
		"",
		projectID,
	)
	if err != nil {
		common.WriteResponse(c, nil, "failed to register", http.StatusBadRequest)
		return
	}

	// insert db
	err = database.UserWrapper.InsertUser(&req)
	if err != nil {
		common.WriteResponse(c, nil, "failed to insert user to database", http.StatusBadRequest)
		return
	}

	accessToken, err := common.GenerateAccess(isPremium)
	if err != nil {
		common.WriteResponse(c, nil, "failed to generate token", http.StatusInternalServerError)
		return
	}
	refreshToken, err := common.GenerateRefresh(isPremium)
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

	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(
		req.RefreshToken,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			return environment.GetSecretKey(), nil
		},
	)

	if err != nil || !token.Valid {
		common.WriteResponse(c, nil, "invalid token", http.StatusBadRequest)
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
