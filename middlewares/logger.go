package middlewares

import (
	"os"
	"log"
	"fmt"
	"time"
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
	//"github.com/gin-contrib/requestid"
)

func Logger(notLogged ...string) gin.HandlerFunc {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknow"
	}

	var skip map[string]struct{}

	if length := len(notLogged); length > 0 {
		skip = make(map[string]struct{}, length)

		for _, p := range notLogged {
			skip[p] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		path := c.Request.URL.Path
		start := time.Now()
		c.Next()
		stop := time.Since(start)
		latency := int(math.Ceil(float64(stop.Nanoseconds()) / 1000000.0))
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		clientUserAgent := c.Request.UserAgent()
		referer := c.Request.Referer()
		dataLength := c.Writer.Size()
		if dataLength < 0 {
			dataLength = 0
		}

		if _, ok := skip[path]; ok {
			return
		}

		if len(c.Errors) > 0 {
			log.Print(c.Errors.ByType(gin.ErrorTypePrivate).String())
		} else {
			msg := fmt.Sprintf("%s - %s \"%s %s\" %d %d \"%s\" \"%s\" (%dms)",
				clientIP,
				hostname,
				c.Request.Method,
				path,
				statusCode,
				dataLength,
				referer,
				clientUserAgent,
				latency,
			)
			if statusCode >= http.StatusInternalServerError {
				log.Print(msg)
			} else if statusCode >= http.StatusBadRequest {
				log.Print(msg)
			} else {
				log.Print(msg)
			}
		}
	}
}
