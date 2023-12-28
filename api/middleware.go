package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"top_100_billboard_golang/environment"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type TokenType int

const (
	Access = iota
	Refresh
)

func GenerateRefresh() (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = time.Now().Add(10 * 365 * 24 * time.Hour).Unix()
	claims["token_type"] = Refresh

	signedToken, err := token.SignedString(environment.GetSecretKey())
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func GenerateAccess() (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = time.Now().Add(1 * time.Hour).Unix()
	claims["token_type"] = Access

	signedToken, err := token.SignedString(environment.GetSecretKey())
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func AccessTokenMiddleware(c *gin.Context) {
	// CLIENT SIDE
	// access + refresh simpen di cache
	// selipin access di "Authorization" header

	// SERVER SIDE
	// check "Authorization" header
	// jenis token & expiration date
	authHeader := c.Request.Header.Get("Authorization")
	// "Bearer {token}"
	accessToken := strings.Replace(authHeader, "Bearer ", "", 1)

	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(environment.GetSecretKey()), nil
	})
	if err != nil {
		writeResponse(c, nil, err.Error(), http.StatusBadRequest)
		c.Abort()
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		writeResponse(c, nil, err.Error(), http.StatusBadRequest)
		c.Abort()
		return
	}

	tokenType, err := strconv.Atoi(fmt.Sprintf("%v", claims["token_type"]))
	if err != nil {
		writeResponse(c, nil, err.Error(), http.StatusBadRequest)
		c.Abort()
		return
	}

	if TokenType(tokenType) != Access {
		writeResponse(c, nil, err.Error(), http.StatusBadRequest)
		c.Abort()
		return
	}

	c.Next()
}
