package common

import "github.com/gin-gonic/gin"

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
