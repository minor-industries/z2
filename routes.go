//go:build !wasm

package main

import (
	"bytes"
	"embed"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/minor-industries/backup/restic"
	"github.com/minor-industries/calendar/gen/go/calendar"
	"github.com/minor-industries/rtgraph"
	"github.com/minor-industries/rtgraph/broker"
	"github.com/minor-industries/rtgraph/database/sqlite"
	"github.com/minor-industries/z2/app"
	handler2 "github.com/minor-industries/z2/app/handler"
	"github.com/minor-industries/z2/backup"
	"github.com/minor-industries/z2/cfg"
	"github.com/minor-industries/z2/frontend"
	"github.com/minor-industries/z2/gen/go/api"
	"github.com/minor-industries/z2/handler"
	"github.com/minor-industries/z2/sync"
	"github.com/minor-industries/z2/variables"
	"github.com/minor-industries/z2/workouts"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"html/template"
	"io/fs"
	"net/http"
	"strconv"
)

//go:embed frontend/templates/*.html
var templatesFS embed.FS

//go:embed frontend/env_gin.js
var envWebJSTemplate []byte

func renderEnvTemplate(data map[string]any) ([]byte, error) {
	tmpl, err := template.New("env.js").Parse(string(envWebJSTemplate))
	if err != nil {
		return nil, err
	}

	var result bytes.Buffer
	if err := tmpl.Execute(&result, data); err != nil {
		return nil, err
	}

	return result.Bytes(), nil
}

func setupRoutes(
	router *gin.Engine,
	opts *cfg.Config,
	graph *rtgraph.Graph,
	br *broker.Broker,
	backends handler.Backends,
	vars *variables.Cache,
	disconnect chan struct{},
) error {
	samples := backends.Samples.(*sqlite.Backend) // TODO

	router.Use(gin.Recovery())
	skipLogging := []string{"/metrics"}
	router.Use(gin.LoggerWithWriter(gin.DefaultWriter, skipLogging...))
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	router.GET("/favicon.ico", func(c *gin.Context) {
		c.Status(204)
	})

	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/z2/pages/home.html")
	})

	graph.SetupServer(router.Group("/rtgraph"))

	tmpl := template.Must(template.New("").ParseFS(templatesFS, "frontend/templates/*"))
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

	envJS, err := renderEnvTemplate(map[string]any{
		"Sync": map[string]any{
			"Host":     opts.Sync.Host,
			"Days":     opts.Sync.Days,
			"Database": opts.Sync.Database,
		},
	})
	if err != nil {
		return errors.Wrap(err, "render env.js")
	}

	subFS, _ := fs.Sub(frontend.FS, "z2")
	httpFS := http.FS(subFS)
	router.GET("/z2/*filepath", func(c *gin.Context) {
		filepath := c.Param("filepath")
		if filepath == "/env.js" {
			c.Data(http.StatusOK, "application/javascript", envJS)
			return
		}

		if opts.StaticPath != "" {
			c.File(opts.StaticPath + filepath)
		} else {
			c.FileFromFS(filepath, httpFS)
		}
	})

	apiHandler := handler2.NewApiServer(backends, vars, disconnect)
	router.Any("/twirp/api.Api/*Method", gin.WrapH(api.NewApiServer(apiHandler, nil)))

	router.POST(
		"/twirp/calendar.Calendar/*Method",
		gin.WrapH(calendar.NewCalendarServer(apiHandler, nil)),
	)

	sse(router, "/events", func(
		c *gin.Context,
		send func(eventName string, data string) error,
	) {
		clientGone := c.Request.Context().Done()

		ch := br.Subscribe()
		defer br.Unsubscribe(ch)

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
					_ = send("play-sound", fmt.Sprintf("%s", msg.Sound))
				}
			}

		}
	})

	sse(router, "/trigger-sync", func(
		c *gin.Context,
		send func(eventName string, data string) error,
	) {
		//TODO: better cleanup, disconnection handling, etc.
		//TODO: only allow one sync to run at once? (or similar)

		defer func() {
			_ = send("close", "close")
		}()

		host := c.Query("host")
		database := c.Query("database")
		daysStr := c.Query("days")

		days, err := strconv.Atoi(daysStr)
		if err != nil {
			_ = send("server-error", "invalid days parameter")
			return
		}

		syncClient := sync.NewClient(host, database)

		err = sync.Sync(samples, syncClient, days, func(s string) {
			_ = send("info", s)
		})

		if err != nil {
			_ = send("server-error", errors.Wrap(err, "sync").Error())
			return
		}

		_ = send("info", "sync complete")
	})

	sse(router, "/trigger-backup", func(
		c *gin.Context,
		send func(eventName string, data string) error,
	) {
		defer func() {
			_ = send("close", "close")
		}()

		backupPath, err := backup.PrepareForBackup(&opts.Backup)
		if err != nil {
			_ = send("server-error", errors.Wrap(err, "prepare for backup").Error())
			return
		}

		if err := restic.Run(&opts.Backup, backupPath, []string{"."}, restic.LogMessages(func(msg string) error {
			_ = send("info", msg)
			return nil
		})); err != nil {
			_ = send("server-error", errors.Wrap(err, "run backup").Error())
			return
		}
	})

	if opts.SyncServer.Enable {
		err := sync.SetupRoutes(router, &opts.SyncServer)
		if err != nil {
			panic(err)
		}
	}

	return nil
}

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
