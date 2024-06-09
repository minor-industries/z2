package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"time"
)

func setupSse(router *gin.Engine) {
	router.GET("/events", func(c *gin.Context) {
		// Set headers for SSE
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")

		// Create a ticker to send events every second
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		// Close connection when client disconnects
		clientGone := c.Request.Context().Done()

		for {
			select {
			case <-clientGone:
				fmt.Println("Client disconnected")
				return
			case t := <-ticker.C:
				// Send an event to the client
				message := fmt.Sprintf("data: Current time is %s\n\n", t.Format(time.RFC3339))
				_, err := c.Writer.Write([]byte(message))
				if err != nil {
					fmt.Println("Error writing to client:", err)
					return
				}

				// Flush the response
				c.Writer.Flush()
			}
		}
	})
}
