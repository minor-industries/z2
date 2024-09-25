package main

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/minor-industries/z2/backup"
	"os"
)

func main() {
	if err := backup.Run(func(msg any) error {
		fmt.Println(spew.Sdump(msg))
		return nil
	}); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
