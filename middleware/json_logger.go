package middleware

import (
	"fmt"
	"time"

	// "github.com/evalphobia/logrus_sentry"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	// "github.com/xin053/ipd/config"
	"github.com/xin053/ipd/utils"
)

// JSONLogMiddleware logs a gin HTTP request in JSON format, with some additional custom key/values
func JSONLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Process Request
		c.Next()

		// Stop timer
		duration := fmt.Sprintf("%fms", utils.GetDurationInMillseconds(start))

		entry := log.WithFields(log.Fields{
			"client_ip": utils.GetClientIP(c),
			"duration":  duration,
			"method":    c.Request.Method,
			"path":      c.Request.RequestURI,
			"status":    c.Writer.Status(),
		})

		statusCode := c.Writer.Status()
		switch {
		case statusCode >= 500:
			entry.Error(c.Errors.String())
		case statusCode >= 400:
			entry.Warn("You should pay attention to this request IP.")
		default:
			entry.Info("")
		}
	}
}

// func init() {
// 	if config.UseSentry {
// 		hook, err := logrus_sentry.NewSentryHook(config.SentryDSN, []log.Level{
// 			log.PanicLevel,
// 			log.FatalLevel,
// 			log.ErrorLevel,
// 		})
// 		hook.Timeout = 5 * time.Second

// 		if err == nil {
// 			log.AddHook(hook)
// 		}
// 	}
// }
