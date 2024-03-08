package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/minor-industries/codelab/cmd/bike/assets"
	handler2 "github.com/minor-industries/codelab/cmd/bike/handler"
	"github.com/minor-industries/codelab/cmd/bike/schema"
	"github.com/minor-industries/platform/common/broker"
	"github.com/pkg/errors"
	"html/template"
	"io/fs"
	"net/http"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
	"time"
)

func serve(
	handler_ *handler2.BleHandler,
	br *broker.Broker,
) error {
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

	r.GET("/data.json", func(c *gin.Context) {
		//query := c.DefaultQuery("id", "1")
		//series, err := strconv.Atoi(query)
		//if err != nil {
		//	_ = c.AbortWithError(400, errors.Wrap(err, "strconv"))
		//	return
		//}

		//data, err := handler.GetSeries(uint16(series))
		//if err != nil {
		//	_ = c.AbortWithError(400, err)
		//	return
		//}

		rows := [][2]any{}
		//for _, d := range data {
		//	// TODO: NaNs for gaps
		//	rows = append(rows, [2]any{d.Timestamp.UnixMilli(), d.Value})
		//}

		c.JSON(200, map[string]any{
			"rows": rows,
		})
	})

	r.GET("/ws", func(c *gin.Context) {
		conn, wsErr := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{
			InsecureSkipVerify: true,
		})
		if wsErr != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, wsErr)
			return
		}

		defer func() {
			_ = conn.Close(websocket.StatusInternalError, "Closed unexpectedly")
		}()

		ctx, cancel := context.WithTimeout(c.Request.Context(), time.Minute)
		defer cancel()
		ctx = conn.CloseRead(ctx)

		msgCh := br.Subscribe()
		defer br.Unsubscribe(msgCh)

		for {
			select {
			case <-ctx.Done():
				conn.Close(websocket.StatusNormalClosure, "")
				return
			case msg := <-msgCh:
				switch m := msg.(type) {
				case *schema.Series:
					switch m.SeriesName {
					case "bike_instant_speed":
						fmt.Println("tick", m.SeriesName, m.Timestamp, m.Value)
						newRows := [][2]any{{
							m.Timestamp.UnixMilli(),
							m.Value,
						}}
						writeErr := wsjson.Write(ctx, conn, map[string]any{
							"rows": newRows,
						})
						if writeErr != nil {
							fmt.Println("error writing to websocket:", writeErr.Error())
							return
						}
					}
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
