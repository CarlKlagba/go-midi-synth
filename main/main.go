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
		startWithNoMidi()
		return
	}
	wp := audio.NewWaveProcessor()

	err := audio.StreamAudio(wp)
	must(err)

	err = audio.StartReadingMidiMessages(wp)
	must(err)

	keyreader.Controls(&wp.AtomicWaveform)

	select {}
}

func startWithNoMidi() {
	wp := audio.NewWaveProcessor()

	err := audio.StreamAudio(wp)
	must(err)

	note := audio.MidiNote{Note: 69, Velocity: 80}
	notes := audio.NotesPlayed{Notes: []audio.MidiNote{note}}
	wp.AtomicPlayedNotes.Store(notes)

	keyreader.Controls(&wp.AtomicWaveform)

	time.Sleep(50 * time.Minute)
	return
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
