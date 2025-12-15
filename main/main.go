package main

import (
	"log"
	"os"

	"github.com/CarlKlagba/go-midi-synth/audio"
	"github.com/CarlKlagba/go-midi-synth/tui"
	tea "github.com/charmbracelet/bubbletea"
)

type progFlags struct {
	noNotesDisplay bool
	noMidi         bool
}

var flags = progFlags{
	noNotesDisplay: false,
	noMidi:         false,
}

func (p *progFlags) flag(progArg []string) {
	for _, a := range progArg {
		if a == "--no-note-display" || a == "-nd" {
			p.noNotesDisplay = true
			log.Println("No Notes Display")
		}
		if a == "--no-midi" || a == "-nm" {
			p.noMidi = true
			log.Println("No Midi")
		}
	}
}

func main() {
	args := os.Args[1:]
	flags.flag(args)
	flags.noNotesDisplay = true // we for it at true until with fix the perf issues

	wp := audio.NewWaveProcessor()
	stream, err := audio.StreamAudio(wp)
	if err != nil {
		log.Fatalf("Failed to start audio stream: %v", err)
	}
	defer audio.CloseAudioStream(stream)

	var midiNotesChan chan []uint8 = nil
	if !flags.noNotesDisplay {
		//Putting the buffer size to 100 seems to fix the issue with clipping but need to look further into it
		midiNotesChan = make(chan []uint8, 100)
		defer close(midiNotesChan)
	}

	if !flags.noMidi {
		err = audio.StartReadingMidiMessages(wp, midiNotesChan)
		if err != nil {
			log.Fatalf("Failed to start midi reading: %v", err)
		}
		defer audio.CloseMidiReader()
	}

	p := tea.NewProgram(tui.InitialModel(wp, midiNotesChan))
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error starting the TUI: %v", err)
	}
}
