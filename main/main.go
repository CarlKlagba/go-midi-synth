package main

import (
	"fmt"
	"github.com/CarlKlagba/go-midi-synth/audio"
	"github.com/CarlKlagba/go-midi-synth/keyreader"
	"github.com/gordonklaus/portaudio"
	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/reader"
	"gitlab.com/gomidi/rtmididrv"
	"log"
	"os"
	"time"
)

const sampleRate = 44100

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

	must(portaudio.Initialize())
	defer portaudio.Terminate()

	wp := audio.NewWaveProcessor()
	stream, err := portaudio.OpenDefaultStream(
		0,
		1,
		sampleRate,
		256,
		wp.ProcessAudio,
	)
	must(err)
	defer stream.Close()
	must(stream.Start())
	defer stream.Stop()

	// Si l'option --no-midi est fournie, on n'utilise pas le MIDI
	if noMidi {
		note := audio.MidiNote{Note: 69, Velocity: 80}
		notes := audio.NotesPlayed{Notes: []audio.MidiNote{note}}
		wp.AtomicPlayedNotes.Store(notes)
		keyreader.Controls(&wp.AtomicWaveform)
		time.Sleep(50 * time.Minute)
		return
	}

	drv, err := rtmididrv.New()
	if err != nil {
		log.Fatal(err)
	}
	defer must(drv.Close())

	ins, err := drv.Ins()
	if err != nil {
		log.Fatal(err)
	}

	if len(ins) == 0 {
		log.Fatal("Aucun port MIDI d'entrée trouvé")
	}

	in := ins[0]
	must(in.Open())

	defer func(in midi.In) {
		must(in.Close())
	}(in)

	rd := reader.New(
		reader.NoLogger(),
		reader.Each(audio.ListenToMidiMessage(&wp.AtomicPlayedNotes)),
	)

	fmt.Println("En attente de messages MIDI...")
	err1 := rd.ListenTo(in)
	must(err1)

	keyreader.Controls(&wp.AtomicWaveform)

	time.Sleep(50 * time.Minute)
}

func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}
