//go:build !wasm

package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/minor-industries/calendar/gen/go/calendar"
	"github.com/minor-industries/rtgraph"
	"github.com/minor-industries/rtgraph/broker"
	"github.com/minor-industries/rtgraph/database/sqlite"
	"github.com/minor-industries/z2/app"
	handler2 "github.com/minor-industries/z2/app/handler"
	"github.com/minor-industries/z2/cfg"
	"github.com/minor-industries/z2/gen/go/api"
	"github.com/minor-industries/z2/handler"
	"github.com/minor-industries/z2/static/dist"
	"github.com/minor-industries/z2/sync"
	"github.com/minor-industries/z2/variables"
	"github.com/minor-industries/z2/workouts"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"html/template"
	"net/http"
	"strconv"
)

func setupRoutes(
	router *gin.Engine,
	opts *cfg.Config,
	graph *rtgraph.Graph,
	br *broker.Broker,
	backends handler.Backends,
	vars *variables.Cache,
) {
	samples := backends.Samples.(*sqlite.Backend) // TODO

	router.Use(gin.Recovery())
	skipLogging := []string{"/metrics"}
	router.Use(gin.LoggerWithWriter(gin.DefaultWriter, skipLogging...))
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	router.GET("/favicon.ico", func(c *gin.Context) {
		c.Status(204)
	})

	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/dist/html/home.html")
	})

	graph.SetupServer(router.Group("/rtgraph"))

	tmpl := template.Must(template.New("").ParseFS(templatesFS, "templates/*"))
	router.SetHTMLTemplate(tmpl)

	router.GET("/workouts.html", func(c *gin.Context) {
		ref := c.Query("ref")

		cfg, ok := app.Configs[ref]
		if !ok {
			_ = c.Error(errors.New("unknown config"))
			return
		}

		data, err := workouts.GenerateData(samples.GetORM(), ref, cfg.PaceMetric)
		if err != nil {
			_ = c.Error(err)
			return
		}
		c.HTML(http.StatusOK, "workouts.html", gin.H{
			"Title": ref + " workouts",
			"Data":  data,
			"Ref":   ref,
		})
	})

	setupSse(br, router)

	if opts.StaticPath != "" {
		router.Static("/dist", opts.StaticPath)
	} else {
		router.StaticFS("/dist", http.FS(dist.FS))
	}

	router.GET("/env.js", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/javascript", envWebJS)
	})

	apiHandler := handler2.NewApiServer(backends, vars)
	router.Any("/twirp/api.Api/*Method", gin.WrapH(api.NewApiServer(apiHandler, nil)))

	router.POST(
		"/twirp/calendar.Calendar/*Method",
		gin.WrapH(calendar.NewCalendarServer(apiHandler, nil)),
	)

	router.GET("/trigger-sync", func(c *gin.Context) {
		//TODO: better cleanup, disconnection handling, etc.
		//TODO: create reusable sse code and use in all sse handlers
		//TODO: only allow one sync to run at once? (or similar)

		host := c.Query("host")
		database := c.Query("database")
		daysStr := c.Query("days")

		info := func(msg string) {
			_, err := c.Writer.Write([]byte("data: " + msg + "\n\n"))
			if err != nil {
				fmt.Println("error writing to client:", err)
				return
			}
			c.Writer.Flush()
		}

		days, err := strconv.Atoi(daysStr)
		if err != nil {
			info("invalid days parameter") // TODO: error when available
			return
		}

		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")

		syncClient := sync.NewClient(host, database)

		err = sync.Sync(samples, syncClient, days, info)
		if err != nil {
			info("sync error: " + err.Error())
		}
	})
}

func setupSse(
	br *broker.Broker,
	router *gin.Engine,
) {
	router.GET("/events", func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")

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
