package audio

import (
	"errors"
	"log"
	"slices"

	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/reader"
	"gitlab.com/gomidi/rtmididrv"
)

type MidiNote struct {
	Note     uint8
	Velocity uint8
	On       bool
}

const midiNoteOn byte = 0x90
const midiNoteOff byte = 0x80

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
		return errors.New("no midi input device found")
	}

	midiIn = ins[0]
	must(midiIn.Open())

	log.Println("MIDI Input Device:", midiIn.String())

	rd := reader.New(
		reader.NoLogger(),
		reader.Each(listenToMidiMessage(wp, notesSender)),
	)

	log.Println("Listening to MIDI messages...")
	err = rd.ListenTo(midiIn)

	if err != nil {
		return err
	}

	return nil
}

func CloseMidiReader() {
	log.Println("Closing Midi Reader...")
	must(driverInstance.Close())
	must(midiIn.Close())
}

func listenToMidiMessage(wp *WaveProcessor, notesChan chan<- []uint8) func(pos *reader.Position, msg midi.Message) {
	return func(pos *reader.Position, msg midi.Message) {
		midiBytes := msg.Raw()
		if len(midiBytes) < 3 {
			return
		}
		channel := midiBytes[0] & 0xF0
		note := midiBytes[1]
		velocity := midiBytes[2]
		//fmt.Printf("Canal: 0x%X, Note: %d, Velocity: %d\n", channel, note, velocity)

		midiNotes := wp.GetNotesPlayed()
		if channel == midiNoteOn && velocity > 0 {
			playedNotes := turnOnNote(note, velocity, *midiNotes)
			wp.SetPlayedNotes(&playedNotes)
		} else if (channel == midiNoteOff) || (channel == midiNoteOn && velocity == 0) {
			playedNotes := turnOffNote(note, *midiNotes)
			wp.SetPlayedNotes(&playedNotes)
		}

		if notesChan != nil {
			notesTemp := *wp.GetNotesPlayed()
			copyNotes := make([]MidiNote, len(notesTemp))
			copy(copyNotes, notesTemp)
			on := notesOn(copyNotes)
			go func() {
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
	newNotes := make([]MidiNote, len(playedNotes))
	copy(newNotes, playedNotes) // trick to avoid race condition.

	for i, n := range newNotes {
		if n.Note == note {
			newNotes[i].Velocity = velocity
			newNotes[i].On = true
			return newNotes
		}
	}
	return append(newNotes, MidiNote{Note: note, Velocity: velocity, On: true})
}

func turnOffNote(note byte, playedNotes []MidiNote) []MidiNote {
	newNotes := make([]MidiNote, len(playedNotes))
	copy(newNotes, playedNotes) // trick to avoid race condition.

	for i, n := range newNotes {
		if n.Note == note {
			newNotes[i].On = false
			break
		}
	}
	return newNotes
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
