package audio

import (
	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/reader"
	"gitlab.com/gomidi/rtmididrv"
	"log"
	"slices"
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

var driverInstance *rtmididrv.Driver
var midiIn midi.In

func StartReadingMidiMessages(wp *WaveProcessor, notesSender chan<- []uint8) error {
	var err error
	driverInstance, err = rtmididrv.New()

	if err != nil {
		return err
	}

	ins, err := driverInstance.Ins()

	if err != nil {
		return err
	}

	if len(ins) == 0 {
		log.Fatal("No MIDI input devices found")
	}

	midiIn = ins[0]
	must(midiIn.Open())

	log.Println("MIDI Input Device:", midiIn.String())

	rd := reader.New(
		reader.NoLogger(),
		reader.Each(listenToMidiMessage(&wp.AtomicPlayedNotes, notesSender)),
	)

	log.Println("Listening to MIDI messages...")
	err = rd.ListenTo(midiIn)

	if err != nil {
		return err
	}

	return nil
}

func CloseMidiReader() {
	must(driverInstance.Close())
	must(midiIn.Close())
}

func listenToMidiMessage(atomicPlayedNotes *atomic.Value, notesChan chan<- []uint8) func(pos *reader.Position, msg midi.Message) {
	return func(pos *reader.Position, msg midi.Message) {
		midiBytes := msg.Raw()
		if len(midiBytes) < 3 {
			return
		}
		canal := midiBytes[0] & 0xF0
		note := midiBytes[1]
		velocity := midiBytes[2]
		//fmt.Printf("Canal: 0x%X, Note: %d, Velocity: %d\n", canal, note, velocity)

		if canal == midiNoteOn && velocity > 0 {
			playedNotes = turnOnNote(note, velocity, playedNotes)
		} else if (canal == midiNoteOff) || (canal == midiNoteOn && velocity == 0) {
			playedNotes = turnOffNote(note, playedNotes)
		}

		n := NotesPlayed{Notes: playedNotes}
		atomicPlayedNotes.Store(n)

		if notesChan != nil {
			go func() {
				on := notesOn(playedNotes)
				slices.Sort(on)
				//fmt.Println("send to chan: ", on)
				notesChan <- on
				//Seem blocking, check if stop blocking with reading
				//fmt.Println("stop blocking ")
			}()
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

func notesOn(midiNotes []MidiNote) []uint8 {
	var nOn []uint8
	for _, n := range midiNotes {
		if n.On {
			nOn = append(nOn, n.Note)
		}
	}
	return nOn
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
