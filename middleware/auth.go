package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/xin053/ipd/config"
)

//AuthRequired Authorization middleware
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" || authHeader != config.AuthKey {
			c.AbortWithStatus(401)
			return
		}
	}
}
