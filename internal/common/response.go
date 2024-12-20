package common

import (
	"fmt"
	"github.com/freeman7728/gorder-v2/common/tracing"
	"github.com/gin-gonic/gin"
	"net/http"
)

type BaseResponse struct {
}

type response struct {
	Errno      int    `json:"errno"`
	Message    string `json:"message"`
	Data       any    `json:"data"`
	TraceID    string `json:"trace_id"`
	TraceIDURL string `json:"trace_id_url"`
}

func (base *BaseResponse) Response(c *gin.Context, err error, data any) {
	if err != nil {
		base.error(c, err)
	} else {
		base.success(c, data)
	}
}

func (base *BaseResponse) success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, response{
		Errno:      0,
		Message:    "success",
		Data:       data,
		TraceID:    tracing.TraceID(c.Request.Context()),
		TraceIDURL: fmt.Sprintf("http://centos2:16686/trace/%s", tracing.TraceID(c.Request.Context())),
	})
}

func (base *BaseResponse) error(c *gin.Context, err error) {
	c.JSON(http.StatusOK, response{
		Errno:      2,
		Message:    err.Error(),
		Data:       nil,
		TraceID:    tracing.TraceID(c.Request.Context()),
		TraceIDURL: fmt.Sprintf("http://centos2:16686/trace/%s", tracing.TraceID(c.Request.Context())),
	})
}
