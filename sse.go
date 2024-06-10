package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/minor-industries/rtgraph/broker"
	"github.com/minor-industries/z2/app"
	"time"
)

func setupSse(
	br *broker.Broker,
	router *gin.Engine,
) {
	router.GET("/events", func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")

		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		clientGone := c.Request.Context().Done()

		ch := br.Subscribe()
		defer br.Unsubscribe(ch)

		sendMsg := func(message string) {
			_, err := c.Writer.Write([]byte("data: " + message + "\n\n"))
			if err != nil {
				fmt.Println("Error writing to client:", err)
				return
			}
			c.Writer.Flush()
		}

		for {
			select {
			case <-clientGone:
				fmt.Println("Client disconnected")
				return
			case t := <-ticker.C:
				sendMsg(
					fmt.Sprintf("Current time is %s", t.Format(time.RFC3339)),
				)
			case m, ok := <-ch:
				if !ok {
					return
				}
				switch msg := m.(type) {
				case *app.PlaySound:
					sendMsg(
						fmt.Sprintf("%s", msg.Sound),
					)
				}
			}

		}
	})
}
