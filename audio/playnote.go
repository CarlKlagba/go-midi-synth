package audio

import (
	"github.com/gordonklaus/portaudio"
)

func PlayNote() error {
	err := portaudio.Initialize()
	if err != nil {
		return err
	}

	wp := NewWaveProcessor()

	stream, err := portaudio.OpenDefaultStream(
		0,
		1,
		sampleRate,
		256,
		wp.ProcessAudio,
	)
	if err != nil {
		return err
	}
	err = stream.Start()
	if err != nil {
		return err
	}

	note := MidiNote{Note: 69, Velocity: 80}
	notes := NotesPlayed{Notes: []MidiNote{note}}
	wp.AtomicPlayedNotes.Store(notes)

	err = stream.Stop()
	if err != nil {
		return err
	}
	err = stream.Close()
	if err != nil {
		return err
	}
	err = portaudio.Terminate()
	if err != nil {
		return err
	}
	return nil
}
