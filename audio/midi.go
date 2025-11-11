package audio

import (
	"fmt"
	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/reader"
	"gitlab.com/gomidi/rtmididrv"
	"log"
	"sync/atomic"
)

type MidiNote struct {
	Note     uint8
	Velocity uint8
	On       bool
}

const midiNoteOn byte = 0x90
const midiNoteOff byte = 0x80

var playedNotes []MidiNote

func StartReadingMidiMessages(wp *WaveProcessor) error {
	drv, err := rtmididrv.New()

	if err != nil {
		return err
	}

	defer must(drv.Close())

	ins, err := drv.Ins()

	if err != nil {
		return err
	}

	if len(ins) == 0 {
		log.Fatal("No MIDI input devices found")
	}

	in := ins[0]
	must(in.Open())
	defer func(in midi.In) {
		must(in.Close())
	}(in)

	log.Println("MIDI Input Device:", in.String())

	rd := reader.New(
		reader.NoLogger(),
		reader.Each(listenToMidiMessage(&wp.AtomicPlayedNotes)),
	)

	log.Println("Listening to MIDI messages...")
	err = rd.ListenTo(in)

	if err != nil {
		return err
	}

	return nil
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

		if canal == midiNoteOn && velocity > 0 {
			playedNotes = turnOnNote(note, velocity, playedNotes)
			n := NotesPlayed{Notes: playedNotes}
			atomicPlayedNotes.Store(n)
		} else if (canal == midiNoteOff) || (canal == midiNoteOn && velocity == 0) {
			playedNotes = turnOffNote(note, playedNotes)
			n := NotesPlayed{Notes: playedNotes}
			atomicPlayedNotes.Store(n)
		}
	}
}

func turnOnNote(note byte, velocity byte, playedNotes []MidiNote) []MidiNote {
	for i, n := range playedNotes {
		if n.Note == note {
			playedNotes[i].Velocity = velocity
			playedNotes[i].On = true
			return playedNotes
		}
	}
	return append(playedNotes, MidiNote{note, velocity, true})
}

func turnOffNote(note byte, playedNotes []MidiNote) []MidiNote {
	for i, n := range playedNotes {
		if n.Note == note {
			playedNotes[i].On = false
			break
		}
	}
	return playedNotes
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
