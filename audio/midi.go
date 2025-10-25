package audio

import (
	"fmt"
	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/reader"
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

func ListenToMidiMessage(atomicPlayedNotes *atomic.Value) func(pos *reader.Position, msg midi.Message) {
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
