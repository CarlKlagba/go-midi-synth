package audio

import (
	"math"
	"sync/atomic"
)

const sampleRate = 44100
const maxAmplitude = 0.8
const maxVelocity = 127
const noteOff = 0

type NotesPlayed struct {
	Notes []MidiNote
}

type Waveform int8

const (
	Sine Waveform = iota
	Square
	Sawtooth
	Triangle
)

type WaveProcessor struct {
	AtomicPlayedNotes atomic.Value
	AtomicWaveform    atomic.Value
	noteToFreq        map[uint8]float64
	noteToPhase       map[uint8]float64
	velocityToAmp     map[uint8]float64
	noteToFadeIn      map[uint8]uint16
	fadeInCount       uint16
	fadeInSample      uint16
	amp               float64
}

func NewWaveProcessor() *WaveProcessor {
	noteToFreq := make(map[uint8]float64, 128)
	noteToPhase := make(map[uint8]float64, 128)
	velocityToAmp := make(map[uint8]float64, 128)
	noteToFadeIn := make(map[uint8]uint16, 128)

	var atomicPlayedNotes atomic.Value
	atomicPlayedNotes.Store(NotesPlayed{Notes: make([]MidiNote, 0, 10)}) //careful we only allocate 10, so we might only be able to play 10notes at the time
	var atomicWaveform atomic.Value
	atomicWaveform.Store(Sine)

	for i := 0; i <= 127; i++ {
		noteToFreq[uint8(i)] = midiNoteToFreq(uint8(i))
	}
	noteToFreq[noteOff] = 0.0

	for i := 0; i <= 127; i++ {
		noteToPhase[uint8(i)] = 0.0
	}

	for i := 0; i <= 127; i++ {
		velocityToAmp[uint8(i)] = midiVelocityToAmplitude(uint8(i))
	}

	for i := 0; i <= 127; i++ {
		noteToFadeIn[uint8(i)] = 0
	}
	return &WaveProcessor{
		atomicPlayedNotes,
		atomicWaveform,
		noteToFreq,
		noteToPhase,
		velocityToAmp,
		noteToFadeIn,
		441, // 10ms de fade-in
		0,
		maxAmplitude,
	}
}
func (w *WaveProcessor) ProcessAudio(out []float32) {
	notesPlayed := w.AtomicPlayedNotes.Load().(NotesPlayed)
	waveform := w.AtomicWaveform.Load().(Waveform)
	for i := range out {
		o := float32(0.0)
		for _, midiNote := range notesPlayed.Notes {
			freq := w.noteToFreq[midiNote.Note]
			phase := w.noteToPhase[midiNote.Note]
			amp := w.velocityToAmp[midiNote.Velocity]

			step := freq / sampleRate

			switch waveform {
			case Sine:
				o += float32(amp * sinWave(phase))
			case Square:
				o += float32(amp * sqrWave(phase))
			case Triangle:
				o += float32(amp * triWave(phase))
			case Sawtooth:
				o += float32(amp * sawWave(phase))
			default:
				o += 0.
			}

			_, w.noteToPhase[midiNote.Note] = math.Modf(phase + step)
		}

		out[i] = o
	}
}

func midiVelocityToAmplitude(velocity uint8) float64 {
	return (float64(velocity) / float64(maxVelocity)) * maxAmplitude
}

func midiNoteToFreq(note uint8) float64 {
	const a4Freq = 440.0
	const a4MidiNote = 69.0
	const numberOfNotes = 12.0
	return a4Freq * math.Pow(2, (float64(note)-a4MidiNote)/numberOfNotes)
}

func sinWave(phase float64) float64 {
	return math.Sin(2 * math.Pi * phase)
}

func sqrWave(phase float64) float64 {
	return math.Copysign(1, math.Sin(2*math.Pi*phase))
}

func triWave(phase float64) float64 {
	return 4*math.Abs(phase-0.5) - 1
}

func sawWave(phase float64) float64 {
	return phase - 0.5
}
