package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/minor-industries/codelab/cmd/bike/assets"
	"github.com/minor-industries/codelab/cmd/bike/database"
	handler2 "github.com/minor-industries/codelab/cmd/bike/handler"
	"github.com/pkg/errors"
	"html/template"
	"io/fs"
	"math"
	"net/http"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
	"strconv"
	"strings"
	"time"
)

func serve(handler *handler2.BleHandler) error {
	r := gin.Default()

	funcs := map[string]any{}

	templ := template.Must(template.New("").Funcs(funcs).ParseFS(assets.FS, "*.html"))
	r.SetHTMLTemplate(templ)

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", map[string]any{})
	})

	r.GET("/favicon.ico", func(c *gin.Context) {
		c.Status(204)
	})

	files(r,
		"dygraph.min.js", "application/javascript",
		"dygraph.css", "text/css",

		"graphs.js", "application/javascript",
	)

	/*
		SELECT
			*
		FROM
			`values` AS v,
			series AS s
		WHERE
			v.series_id = s.id
			AND series_id = 3
		ORDER BY
			timestamp ASC
	*/

	r.GET("/data.json", func(c *gin.Context) {
		query := c.DefaultQuery("id", "1")
		series, err := strconv.Atoi(query)
		if err != nil {
			_ = c.AbortWithError(400, errors.Wrap(err, "strconv"))
			return
		}

		data, err := handler.GetSeries(uint16(series))
		if err != nil {
			_ = c.AbortWithError(400, err)
			return
		}

		var rows [][2]any
		for _, d := range data {
			// TODO: NaNs for gaps
			rows = append(rows, [2]any{d.Timestamp.UnixMilli(), d.Value})
		}

		c.JSON(200, map[string]any{
			"rows": rows,
		})
	})

	r.GET("/data.csv", func(c *gin.Context) {
		query := c.DefaultQuery("id", "1")
		series, err := strconv.Atoi(query)
		if err != nil {
			_ = c.AbortWithError(400, errors.Wrap(err, "strconv"))
			return
		}

		data, err := handler.GetSeries(uint16(series))
		if err != nil {
			_ = c.AbortWithError(400, err)
			return
		}

		lines := []string{"timestamp,value"}

		for i, d := range data {
			if i > 0 {
				d0 := data[i-1]
				dt := d.Timestamp.Sub(d0.Timestamp)
				if dt > time.Minute {
					lines = append(lines, fmt.Sprintf("%s,NaN",
						d0.Timestamp.Format("2006/01/02 15:04:05"),
					))
				}
			}

			lines = append(lines, fmt.Sprintf("%s,%f",
				d.Timestamp.Format("2006/01/02 15:04:05"),
				d.Value,
			))
		}

		content := []byte(strings.Join(lines, "\n"))

		c.Writer.Header().Set("Content-Type", "text/plain")
		c.Status(200)
		_, _ = c.Writer.Write(content)
	})

	t0 := time.Now()
	r.GET("/ws", func(c *gin.Context) {
		conn, wsErr := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{
			InsecureSkipVerify: true,
		})
		if wsErr != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, wsErr)
			return
		}

		var v database.Value
		tx := handler.Db.Order("timestamp desc").Limit(1).Find(&v)
		if tx.Error != nil {
			panic(errors.Wrap(tx.Error, "find")) // TODO
		}

		defer func() {
			_ = conn.Close(websocket.StatusInternalError, "Closed unexpectedly")
		}()

		ctx, cancel := context.WithTimeout(c.Request.Context(), time.Minute)
		defer cancel()
		ctx = conn.CloseRead(ctx)

		tick := time.NewTicker(250 * time.Millisecond)
		defer tick.Stop()

		for {
			select {
			case <-ctx.Done():
				conn.Close(websocket.StatusNormalClosure, "")
				return
			case <-tick.C:
				dt := time.Now().Sub(t0)
				t := v.Timestamp.Add(dt)

				newRows := [][2]any{{
					t.UnixMilli(),
					5*math.Sin(dt.Seconds()/10.0) + 30,
				}}
				writeErr := wsjson.Write(ctx, conn, map[string]any{
					"rows": newRows,
				})
				if writeErr != nil {
					return
				}
			}
		}
	})

	if err := r.Run("0.0.0.0:8000"); err != nil {
		return errors.Wrap(err, "run")
	}

	return nil
}

func files(r *gin.Engine, files ...string) {
	for i := 0; i < len(files); i += 2 {
		name := files[i]
		ct := files[i+1]
		r.GET("/"+name, func(c *gin.Context) {
			header := c.Writer.Header()
			header["Content-Type"] = []string{ct}
			content, err := fs.ReadFile(assets.FS, name)
			if err != nil {
				c.Status(404)
				return
			}
			_, _ = c.Writer.Write(content)
		})
	}
}
