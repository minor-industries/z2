package main

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
)

//go:embed sounds/*.wav
var audioFiles embed.FS

func playAudio(fileName string) {
	fmt.Println("play", fileName)

	// Open the audio file from the embed.FS
	audioFile, err := audioFiles.Open(fileName)
	if err != nil {
		log.Fatalf("Failed to open audio file: %v", err)
	}
	defer audioFile.Close()

	// Read the audio file into a buffer
	var buf bytes.Buffer
	_, err = io.Copy(&buf, audioFile)
	if err != nil {
		log.Fatalf("Failed to read audio file: %v", err)
	}

	stream, format, err := wav.Decode(&buf)

	if err != nil {
		log.Fatalf("Failed to decode audio file: %v", err)
	}
	defer stream.Close()

	// Initialize the speaker
	err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	if err != nil {
		log.Fatalf("Failed to initialize speaker: %v", err)
	}

	// Play the audio file
	done := make(chan bool)
	speaker.Play(beep.Seq(stream, beep.Callback(func() {
		done <- true
	})))

	// Wait until playback is finished
	<-done
}

func (app *App) PlaySounds() {
	duration := 5 * time.Second
	ticker := time.NewTicker(duration)
	state := "undefined"
	t0 := time.Now()

	for {
		select {
		case <-ticker.C:
			fmt.Println(time.Now().Sub(t0).Seconds(), "still in", state)
			switch state {
			case "too_slow":
				playAudio("sounds/Computer Error Alert-SoundBible.com-783113881.wav")
			}
		case sc := <-app.StateChanges:
			ticker.Stop()
			ticker = time.NewTicker(duration)
			fmt.Println(time.Now().Sub(t0).Seconds(), sc.From, "->", sc.To)
			state = sc.To

			switch state {
			case "too_slow":
				playAudio("sounds/Computer Error Alert-SoundBible.com-783113881.wav")
			}
		}
	}
}
