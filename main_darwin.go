package main

import (
	"fmt"
	"github.com/minor-industries/codelab/cmd/z2/handler"
	"github.com/minor-industries/codelab/cmd/z2/handler/rower"
	"github.com/minor-industries/codelab/cmd/z2/source"
)

func getSource(name string) (string, source.Source) {
	switch name {
	case "bike":
		return "5f953f6e-382c-39f7-c5ce-643b6141c967", &handler.BikeSource{}
	case "rower":
		return "41f94de4-2cb8-a329-d685-2644a65213ff", rower.NewRowerSource()
	default:
		panic(fmt.Errorf("unknown source: %s", name))
	}
}
