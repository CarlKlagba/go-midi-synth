package audio

import (
	"math"
	"sync/atomic"
)

const sampleRate = 44100
const amplitude = 0.5
const noteOff = 0

type NotesPlayed struct {
	Notes []uint8
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
	noteToFreq        map[uint32]float64
	noteToPhase       map[uint32]float64
	noteToFadeIn      map[uint32]uint16
	fadeInCount       uint16
	fadeInSample      uint16
	amp               float64
}

func NewWaveProcessor() *WaveProcessor {
	noteToFreq := make(map[uint32]float64)
	noteToPhase := make(map[uint32]float64)
	noteToFadeIn := make(map[uint32]uint16)

	var atomicPlayedNotes atomic.Value
	atomicPlayedNotes.Store(NotesPlayed{Notes: make([]uint8, 0, 10)})
	var atomicWaveform atomic.Value
	atomicWaveform.Store(Sine)

	// Initialiser les fréquences des Notes MIDI
	for i := 0; i <= 127; i++ {
		noteToFreq[uint32(i)] = midiNoteToFreq(uint8(i))
	}
	noteToFreq[noteOff] = 0.0

	// Initialiser les phases des Notes MIDI
	for i := 0; i <= 127; i++ {
		noteToPhase[uint32(i)] = 0.0
	}

	// Initialiser les fades-in des Notes MIDI
	for i := 0; i <= 127; i++ {
		noteToFadeIn[uint32(i)] = 0
	}
	return &WaveProcessor{
		atomicPlayedNotes,
		atomicWaveform,
		noteToFreq,
		noteToPhase,
		noteToFadeIn,
		441, // 10ms de fade-in
		0,
		amplitude,
	}
}
func midiNoteToFreq(note uint8) float64 {
	return 440.0 * math.Pow(2, (float64(note)-69)/12)
}

func (w *WaveProcessor) ProcessAudio(out []float32) {
	notesPlayed := w.AtomicPlayedNotes.Load().(NotesPlayed)
	waveform := w.AtomicWaveform.Load().(Waveform)
	for i := range out {
		if w.fadeInCount < w.fadeInSample {
			w.amp = amplitude * float64(w.fadeInCount) / float64(w.fadeInSample)
			w.fadeInCount++
		}

		o := float32(0.0)
		for _, v := range notesPlayed.Notes {
			freq := w.noteToFreq[uint32(v)]
			phase := w.noteToPhase[uint32(v)]
			step := freq / sampleRate

			switch waveform {
			case Sine:
				o += float32(w.amp * sinWave(phase))
			case Square:
				o += float32(w.amp * sqrWave(phase))
			case Triangle:
				o += float32(w.amp * triWave(phase))
			case Sawtooth:
				o += float32(w.amp * sawWave(phase))
			default:
				o += 0.
			}

			_, w.noteToPhase[uint32(v)] = math.Modf(phase + step)
		}

		out[i] = o
	}
}

func sinWave(phase float64) float64 {
	return math.Sin(2 * math.Pi * phase)
}

func sqrWave(phase float64) float64 {
	return math.Copysign(1, math.Sin(2*math.Pi*phase))
}

func triWave(phase float64) float64 {
	return math.Abs(phase-0.5) - 1
}

func sawWave(phase float64) float64 {
	return phase - 0.5
}
