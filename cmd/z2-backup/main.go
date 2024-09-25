package main

import (
	"fmt"
	"github.com/minor-industries/z2/backup"
	"os"
)

func main() {
	if err := backup.Run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
