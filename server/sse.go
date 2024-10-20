package server

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func sse(
	router *gin.Engine,
	path string,
	handler func(
		c *gin.Context,
		send func(eventName, data string) error,
	),
) {
	router.GET(path, func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")

		send := func(eventName, data string) error {
			if eventName != "" {
				_, err := c.Writer.Write([]byte("event: " + eventName + "\n"))
				if err != nil {
					return errors.Wrap(err, "write event")
				}
			}

			_, err := c.Writer.Write([]byte("data: " + data + "\n\n"))
			if err != nil {
				if err != nil {
					return errors.Wrap(err, "write data")
				}
			}
			c.Writer.Flush()
			return nil
		}

		handler(c, send)
	})
}
