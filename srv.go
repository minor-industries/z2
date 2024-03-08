package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/minor-industries/codelab/cmd/bike/assets"
	handler2 "github.com/minor-industries/codelab/cmd/bike/handler"
	"github.com/pkg/errors"
	"html/template"
	"io/fs"
	"net/http"
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

	files(r,
		"dygraph.min.js", "application/javascript",
		"dygraph.css", "text/css",
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

	r.GET("/data.csv", func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "text/plain")
		c.Status(200)

		query := c.DefaultQuery("id", "1")
		series, err := strconv.Atoi(query)
		if err != nil {
			_ = c.AbortWithError(400, errors.Wrap(err, "strconv"))
			return
		}

		lines := []string{"timestamp,value"}
		data, err := handler.GetSeries(uint16(series))
		if err != nil {
			_ = c.Error(err)
			return
		}

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

		_, _ = c.Writer.Write(content)
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
