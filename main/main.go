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
	"sync/atomic"
	"time"
)

const sampleRate = 44100

const midiNoteOn byte = 0x90
const midiNoteOff byte = 0x80

var (
	playedNotes []audio.MidiNote
	noMidi      = false
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
		reader.Each(listenToMidiMessage(&wp.AtomicPlayedNotes)),
	)

	fmt.Println("En attente de messages MIDI...")
	err1 := rd.ListenTo(in)
	must(err1)

	keyreader.Controls(&wp.AtomicWaveform)

	time.Sleep(50 * time.Minute)
}

func listenToMidiMessage(atomicPlayedNotes *atomic.Value) func(pos *reader.Position, msg midi.Message) {
	return func(pos *reader.Position, msg midi.Message) {
		midiBytes := msg.Raw()
		if len(midiBytes) < 3 {
			return
		}
		canal := midiBytes[0] & 0xF0
		note := midiBytes[1]
		velocity := midiBytes[2]
		fmt.Printf("Canal: 0x%X, Note: %d, Velocity: %d\n", canal, note, velocity)

		if canal == midiNoteOn && velocity > 0 { // Note ON
			playedNotes = append(playedNotes, audio.MidiNote{note, velocity})
			n := audio.NotesPlayed{Notes: playedNotes}
			atomicPlayedNotes.Store(n)
		} else if (canal == midiNoteOff) || (canal == midiNoteOn && velocity == 0) { // Note OFF
			for i, n := range playedNotes {
				if n.Note == note {
					playedNotes = append(playedNotes[:i], playedNotes[i+1:]...)
					break
				}
			}
			n := audio.NotesPlayed{Notes: playedNotes}
			atomicPlayedNotes.Store(n)
		}
	}
}

func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}
