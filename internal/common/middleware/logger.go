package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func StructureLog(entry *logrus.Entry) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
