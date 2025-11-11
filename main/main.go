package main

import (
	"github.com/CarlKlagba/go-midi-synth/audio"
	"github.com/CarlKlagba/go-midi-synth/keyreader"
	"log"
	"os"
	"time"
)

var (
	noMidi = false
)

func main() {

	for _, arg := range os.Args[1:] {
		if arg == "--no-midi" {
			noMidi = true
			break
		}
	}

	if noMidi {
		wp := audio.NewWaveProcessor()

		stream, err := audio.StreamAudio(wp)
		must(err)
		defer audio.CloseAudioStream(stream)

		note := audio.MidiNote{Note: 69, Velocity: 80, On: true}
		notes := audio.NotesPlayed{Notes: []audio.MidiNote{note}}
		wp.AtomicPlayedNotes.Store(notes)

		keyreader.Controls(&wp.AtomicWaveform)

		time.Sleep(50 * time.Minute)
		return
	}

	wp := audio.NewWaveProcessor()

	stream, err := audio.StreamAudio(wp)
	must(err)
	defer audio.CloseAudioStream(stream)

	err = audio.StartReadingMidiMessages(wp)
	must(err)
	defer audio.CloseMidiReader()

	keyreader.Controls(&wp.AtomicWaveform)

	select {}
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
