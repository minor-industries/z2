package main

import (
	"fmt"
	webview "github.com/webview/webview_go"
)

func main() {
	debug := true
	w := webview.New(debug)
	defer w.Destroy()
	w.SetTitle("z2")
	w.SetSize(800, 600, webview.HintNone)
	w.Navigate("http://localhost:8000")
	w.Run()
	fmt.Println("here")
}
