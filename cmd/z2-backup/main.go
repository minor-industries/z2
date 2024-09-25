package main

import (
	"fmt"
	"github.com/minor-industries/z2/backup"
	"math"
	"os"
)

func main() {
	lastQuantum := math.MaxFloat64

	if err := backup.Run(func(msg any) error {
		switch msg := msg.(type) {
		case backup.ResticStatus:
			if msg.PercentDone < lastQuantum {
				lastQuantum = 0.0
			}

			currentQuantum := float64(int(msg.PercentDone*10)) / 10.0

			if currentQuantum > lastQuantum {
				lastQuantum = currentQuantum
				fmt.Printf("Progress: %.1f%%\n", msg.PercentDone*100)
			}
		case backup.ResticSummary:
			fmt.Println("Backup done!")
		default:
			fmt.Println("Unknown message type")
		}
		return nil
	}); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
