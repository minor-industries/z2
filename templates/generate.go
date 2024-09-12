package main

//go:generate go run generate.go

import (
	"bytes"
	"embed"
	"html/template"
	"os"
	"path/filepath"
)

//go:embed *.html
var FS embed.FS

type file struct {
	name string
	args any
}

var files = []file{
	{
		name: "bike.html",
		args: map[string]any{
			"Title": "Bike",
		},
	},
	{
		name: "rower.html",
		args: map[string]any{
			"Title": "Rower",
		},
	},
	{
		name: "bike-presets.html",
		args: map[string]any{
			"Title": "Bike Presets",
		},
	},
	{
		name: "data-bike.html",
		args: map[string]any{
			"Title": "Bike Data",
		},
	},
	{
		name: "data-rower.html",
		args: map[string]any{
			"Title": "Rower Data",
		},
	},
	{
		name: "data-hrm.html",
		args: map[string]any{
			"Title": "HRM data",
		},
	},
	{
		name: "home.html",
		args: map[string]any{},
	},
	{
		name: "calendar.html",
		args: map[string]any{},
	},
	// workouts not ready
}

func main() {
	outdir := filepath.Join("..", "static", "dist", "html")
	tmpl := template.Must(template.ParseFS(FS, "*.html"))

	for _, f := range files {
		lookup := tmpl.Lookup(f.name)

		buf := bytes.NewBuffer(nil)
		err := lookup.Execute(buf, f.args)
		noErr(err)

		out := filepath.Join(outdir, f.name)
		noErr(os.WriteFile(out, buf.Bytes(), 0644))
	}
}

func noErr(err error) {
	if err != nil {
		panic(err)
	}
}
