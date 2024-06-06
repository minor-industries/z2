package main

import (
	"bytes"
	"embed"
	"fmt"
	"github.com/ebitengine/oto/v3"
	"github.com/hajimehoshi/go-mp3"
	"github.com/pkg/errors"
	"io/fs"
	"sync"
	"time"
)

//go:embed sounds/*.mp3
var audioFiles embed.FS

var contextOnce sync.Once
var otoContext *oto.Context

func playAudio(fileName string) error {
	fmt.Println("play", fileName)

	fileBytes, err := fs.ReadFile(audioFiles, fileName)
	if err != nil {
		return errors.Wrap(err, "read audio file")
	}

	contextOnce.Do(func() {
		var ready chan struct{}
		otoContext, ready, err = oto.NewContext(&oto.NewContextOptions{
			SampleRate:   44100,
			ChannelCount: 2,
			Format:       oto.FormatSignedInt16LE,
		})
		if err != nil {
			return
		}
		<-ready
	})
	if err != nil {
		return fmt.Errorf("failed to create Oto context: %w", err)
	}

	decoder, err := mp3.NewDecoder(bytes.NewBuffer(fileBytes))
	if err != nil {
		return fmt.Errorf("failed to decode MP3: %w", err)
	}

	player := otoContext.NewPlayer(decoder)
	defer player.Close()

	player.Play()

	for player.IsPlaying() {
		time.Sleep(time.Millisecond)
	}

	return nil
}

func (app *App) PlaySounds() {
	duration := 5 * time.Second
	ticker := time.NewTicker(duration)
	state := "undefined"

	for {
		select {
		case <-ticker.C:
			switch state {
			case "too_slow":
				must(playAudio("sounds/slow.mp3"))
			case "too_fast":
				must(playAudio("sounds/fast.mp3"))
			}

		case sc := <-app.StateChanges:
			ticker.Stop()
			ticker = time.NewTicker(duration)
			state = sc.To

			switch state {
			case "too_slow":
				must(playAudio("sounds/slow.mp3"))
			case "too_fast":
				must(playAudio("sounds/fast.mp3"))
			case "fairway":
				must(playAudio("sounds/fairway.mp3"))
			}
		}
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
