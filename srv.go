package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/minor-industries/codelab/cmd/bike/assets"
	"github.com/minor-industries/codelab/cmd/bike/database"
	"github.com/minor-industries/codelab/cmd/bike/schema"
	"github.com/minor-industries/platform/common/broker"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gorm.io/gorm"
	"html/template"
	"io/fs"
	"net/http"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
	"time"
)

func serve(
	db *gorm.DB,
	br *broker.Broker,
	allSeries map[string]*database.Series,
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

	r.GET("/ws", func(c *gin.Context) {
		ctx := c.Request.Context()

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

		_, reqBytes, err := conn.Read(ctx)
		if wsErr != nil {
			fmt.Println("ws read error", err.Error())
			return
		}
		conn.CloseRead(ctx)

		var req map[string]any
		err = json.Unmarshal(reqBytes, &req)
		if wsErr != nil {
			fmt.Println("ws error", errors.Wrap(err, "unmarshal json"))
			return
		}

		subscribed := req["series"].(string)

		if err := sendInitialData(
			ctx,
			db,
			allSeries,
			conn,
			subscribed,
		); err != nil {
			err := errors.Wrap(err, "send initial data")
			fmt.Println("error", err.Error())
			_ = wsjson.Write(ctx, conn, map[string]any{
				"error": err.Error(),
			})
			return
		}

		msgCh := br.Subscribe()
		defer br.Unsubscribe(msgCh)

		for msg := range msgCh {
			switch m := msg.(type) {
			case *schema.Series:
				if m.SeriesName == subscribed {
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
	})

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	if err := r.Run("0.0.0.0:8000"); err != nil {
		return errors.Wrap(err, "run")
	}

	return nil
}

func sendInitialData(
	ctx context.Context,
	db *gorm.DB,
	allSeries map[string]*database.Series,
	conn *websocket.Conn,
	subscribed string,
) error {
	s, ok := allSeries[subscribed]
	if !ok {
		return errors.New("unknown series")
	}

	data, err := database.LoadData(db, s.ID)
	if err != nil {
		return errors.Wrap(err, "load data")
	}

	rows := [][2]any{}
	for _, d := range data {
		// TODO: NaNs for gaps
		rows = append(rows, [2]any{
			d.Timestamp.UnixMilli(),
			d.Value,
		})
	}

	if err := wsjson.Write(ctx, conn, map[string]any{
		"now": time.Now().UnixMilli(),
	}); err != nil {
		return errors.Wrap(err, "write timestamp")
	}

	if err := wsjson.Write(ctx, conn, map[string]any{
		"initial_data": rows,
		"now":          time.Now().UnixMilli(),
	}); err != nil {
		return errors.Wrap(err, "write initial data")
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
