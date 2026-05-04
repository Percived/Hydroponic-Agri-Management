package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Envelope struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
	RequestID string      `json:"request_id"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Envelope{
		Code:      0,
		Message:   "ok",
		Data:      data,
		RequestID: c.GetString("request_id"),
	})
}

func Error(c *gin.Context, httpStatus int, code int, message string, data interface{}) {
	c.JSON(httpStatus, Envelope{
		Code:      code,
		Message:   message,
		Data:      data,
		RequestID: c.GetString("request_id"),
	})
}
