package common

import (
	"fmt"
	"strings"
	"top_100_billboard_golang/environment"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type genericResponse struct {
	Data       interface{} `json:"data"`
	Message    string      `json:"message"`
	StatusCode int         `json:"status_code"`
}

func WriteResponse(
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

func ParseJWT(c *gin.Context) (jwt.MapClaims, error) {
	authHeader := c.Request.Header.Get("Authorization")
	tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			return environment.GetSecretKey(), nil
		},
	)
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}
