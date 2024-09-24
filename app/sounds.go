package app

import (
	"bytes"
	"fmt"
	"github.com/ebitengine/oto/v3"
	"github.com/hajimehoshi/go-mp3"
	"github.com/minor-industries/z2/frontend/dist"
	"github.com/pkg/errors"
	"io/fs"
	"sync"
	"time"
)

var contextOnce sync.Once
var otoContext *oto.Context

func (app *App) playAudio(name string) error {
	switch app.audioPlayer {
	case "browser":
		app.broker.Publish(&PlaySound{Sound: name})
		return nil
	case "backend":
	// pass
	default:
		return fmt.Errorf("unknown audio player")
	}

	fileName := fmt.Sprintf("sounds/%s.mp3", name)

	fmt.Println("play", fileName)

	fileBytes, err := fs.ReadFile(dist.FS, fileName)
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
				must(app.playAudio("slow"))
			case "too_fast":
				must(app.playAudio("fast"))
			}

		case sc := <-app.stateChanges:
			ticker.Stop()
			ticker = time.NewTicker(duration)
			state = sc.To

			switch state {
			case "too_slow":
				must(app.playAudio("slow"))
			case "too_fast":
				must(app.playAudio("fast"))
			case "fairway":
				must(app.playAudio("fairway"))
			}
		}
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
