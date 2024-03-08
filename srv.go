package main

import (
	"github.com/gin-gonic/gin"
	"github.com/minor-industries/codelab/cmd/bike/assets"
	handler2 "github.com/minor-industries/codelab/cmd/bike/handler"
	"github.com/pkg/errors"
	"html/template"
	"io/fs"
	"net/http"
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
