package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/minor-industries/codelab/cmd/z2/handler"
	"github.com/minor-industries/codelab/cmd/z2/source"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func getSource(name string) (string, source.Source) {
	switch name {
	case "bike":
		return "FC:38:34:32:0D:69", &handler.BikeSource{}
	default:
		panic(fmt.Errorf("unknown source: %s", name))
	}
}
