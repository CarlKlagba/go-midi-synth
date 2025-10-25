package audio

import (
	"testing"
)

/*
 *
 * go test -bench=. -benchtime=20s -benchmem
 *
 */

func Benchmark_ProcessAudio_Sine(b *testing.B) {
	wp := NewWaveProcessor()
	wp.AtomicWaveform.Store(Sine)
	note := MidiNote{Note: 60, Velocity: 100, On: true}
	wp.AtomicPlayedNotes.Store(NotesPlayed{Notes: []MidiNote{note}})

	out := make([]float32, 256)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wp.ProcessAudio(out)
	}
}

func Benchmark_ProcessAudio_Square(b *testing.B) {
	wp := NewWaveProcessor()
	wp.AtomicWaveform.Store(Square)
	note := MidiNote{Note: 60, Velocity: 100, On: true}
	wp.AtomicPlayedNotes.Store(NotesPlayed{Notes: []MidiNote{note}})

	out := make([]float32, 256)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wp.ProcessAudio(out)
	}
}

func Benchmark_ProcessAudio_Triangle(b *testing.B) {
	wp := NewWaveProcessor()
	wp.AtomicWaveform.Store(Triangle)
	note := MidiNote{Note: 60, Velocity: 100, On: true}
	wp.AtomicPlayedNotes.Store(NotesPlayed{Notes: []MidiNote{note}})

	out := make([]float32, 256)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wp.ProcessAudio(out)
	}
}

func Benchmark_ProcessAudio_Sawtooth(b *testing.B) {
	wp := NewWaveProcessor()
	wp.AtomicWaveform.Store(Sawtooth)
	note := MidiNote{Note: 60, Velocity: 100, On: true}
	wp.AtomicPlayedNotes.Store(NotesPlayed{Notes: []MidiNote{note}})

	out := make([]float32, 256)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wp.ProcessAudio(out)
	}
}

func Benchmark_ProcessAudio_Sine_Note_Off(b *testing.B) {
	wp := NewWaveProcessor()
	wp.AtomicWaveform.Store(Sine)
	note := MidiNote{Note: 60, Velocity: 100, On: false}
	wp.AtomicPlayedNotes.Store(NotesPlayed{Notes: []MidiNote{note}})

	out := make([]float32, 256)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wp.ProcessAudio(out)
	}
}

func Benchmark_ProcessAudio_Mutate_Shared_Data(b *testing.B) {
	wp := NewWaveProcessor()
	wp.AtomicWaveform.Store(Sine)
	note := MidiNote{Note: 60, Velocity: 100, On: false}
	wp.AtomicPlayedNotes.Store(NotesPlayed{Notes: []MidiNote{note}})

	out := make([]float32, 256)

	b.ResetTimer()

	go func() {
		var notes = make([]MidiNote, 127)
		for i := 2; i < b.N; i++ {
			n := uint8(127 % i)
			notes = addNote(n, 100, notes)
			notes = offNote(n-1, notes)
			wp.AtomicPlayedNotes.Store(
				NotesPlayed{
					Notes: notes,
				},
			)
		}
	}()

	for i := 0; i < b.N; i++ {
		wp.ProcessAudio(out)
	}
}

func addNote(note uint8, velocity uint8, notes []MidiNote) []MidiNote {
	for i, n := range notes {
		if n.Note == note {
			notes[i].Velocity = velocity
			notes[i].On = true
			return notes
		}
	}
	return append(notes, MidiNote{note, velocity, true})
}

func offNote(note uint8, playedNotes []MidiNote) []MidiNote {
	for i, n := range playedNotes {
		if n.Note == note {
			playedNotes[i].On = false
			break
		}
	}
	return playedNotes
}
