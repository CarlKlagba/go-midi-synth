package audio

import (
	"math"
	"sync/atomic"
)

const sampleRate = 44100
const initialAmplitude = 0.8
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
	AtomicPlayedNotes  atomic.Value
	AtomicWaveform     atomic.Value
	atomicMaxAmp       atomic.Value
	atomicReleaseCount atomic.Value
	noteToFreq         map[uint8]float64
	noteToPhase        map[uint8]float64
	velocityToAmp      map[uint8]float64
	noteToAttack       map[uint8]uint16
	noteToRelease      map[uint8]uint16
	attackCount        uint16
}

func NewWaveProcessor() *WaveProcessor {
	noteToFreq := make(map[uint8]float64, 128)
	noteToPhase := make(map[uint8]float64, 128)
	velocityToAmp := make(map[uint8]float64, 128)
	noteToAttack := make(map[uint8]uint16, 128)
	noteToRelease := make(map[uint8]uint16, 128)
	releaseCount := uint16(4410) // A mettre en paramettre

	var atomicPlayedNotes atomic.Value
	atomicPlayedNotes.Store(NotesPlayed{Notes: make([]MidiNote, 0, 10)}) //careful we only allocate 10, so we might only be able to play 10notes at the time
	var atomicWaveform atomic.Value
	atomicWaveform.Store(Sine)
	var atomicMaxAmp atomic.Value
	atomicMaxAmp.Store(initialAmplitude)
	var atomicReleaseCount atomic.Value
	atomicReleaseCount.Store(releaseCount)

	for i := 0; i <= 127; i++ {
		noteToFreq[uint8(i)] = midiNoteToFreq(uint8(i))
	}
	noteToFreq[noteOff] = 0.0

	for i := 0; i <= 127; i++ {
		noteToPhase[uint8(i)] = 0.0
	}

	for i := 0; i <= 127; i++ {
		velocityToAmp[uint8(i)] = midiVelocityToAmplitude(uint8(i)) * initialAmplitude
	}

	for i := 0; i <= 127; i++ {
		noteToAttack[uint8(i)] = 0
	}

	for i := 0; i <= 127; i++ {
		noteToRelease[uint8(i)] = releaseCount
	}

	return &WaveProcessor{
		atomicPlayedNotes,
		atomicWaveform,
		atomicMaxAmp,
		atomicReleaseCount,
		noteToFreq,
		noteToPhase,
		velocityToAmp,
		noteToAttack,
		noteToRelease,
		4410, // 10ms de fade-in
	}
}

// ProcessAudio fills the out buffer with audio samples
func (w *WaveProcessor) ProcessAudio(out []float32) {
	notesPlayed := w.AtomicPlayedNotes.Load().(NotesPlayed)
	waveform := w.AtomicWaveform.Load().(Waveform)
	maxAmp := w.atomicMaxAmp.Load().(float64)
	releaseCount := w.atomicReleaseCount.Load().(uint16)

	for i := range out {
		o := float32(0.0)
		for _, midiNote := range notesPlayed.Notes {
			if !midiNote.On && w.noteToRelease[midiNote.Note] == 0 {
				//TODO: remove note when note off after release completed
				continue
			}
			if midiNote.On && w.noteToRelease[midiNote.Note] != releaseCount {
				w.noteToRelease[midiNote.Note] = releaseCount
			}
			if !midiNote.On && w.noteToAttack[midiNote.Note] != 0 {
				w.noteToAttack[midiNote.Note] = 0
			}

			freq := w.noteToFreq[midiNote.Note]
			phase := w.noteToPhase[midiNote.Note]
			amp := w.velocityToAmp[midiNote.Velocity] * maxAmp

			if midiNote.On && w.noteToAttack[midiNote.Note] < w.attackCount {
				amp = amp * (float64(w.noteToAttack[midiNote.Note]) / float64(w.attackCount))
				w.noteToAttack[midiNote.Note]++
			}

			if !midiNote.On && w.noteToRelease[midiNote.Note] > 0 {
				amp = amp * (float64(w.noteToRelease[midiNote.Note]) / float64(releaseCount))
				w.noteToRelease[midiNote.Note]--
			}

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

			step := freq / sampleRate
			_, w.noteToPhase[midiNote.Note] = math.Modf(phase + step)
		}
		out[i] = o
	}
}

func (w *WaveProcessor) SetVolume(volume float64) {
	if volume < 0.0 {
		volume = 0.0
	}
	if volume > 1.0 {
		volume = 1.0
	}
	w.atomicMaxAmp.Store(volume)

}

func (w *WaveProcessor) GetVolume() float64 {
	return w.atomicMaxAmp.Load().(float64)
}

func (w *WaveProcessor) SetReleaseTime(millitsec float64) {
	if millitsec < 5.0 {
		millitsec = 5.0
	}
	if millitsec > 100.0 {
		millitsec = 100.0
	}
	// 441 == 1ms ??
	val := uint16(millitsec * 441)
	w.atomicReleaseCount.Store(val)
}

func (w *WaveProcessor) GetReleaseTime() float64 {
	r := w.atomicReleaseCount.Load().(uint16)
	return float64(r) / 441.0
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

func midiVelocityToAmplitude(velocity uint8) float64 {
	return float64(velocity) / float64(maxVelocity)
}

func midiNoteToFreq(note uint8) float64 {
	const a4Freq = 440.0
	const a4MidiNote = 69.0
	const numberOfNotes = 12.0
	return a4Freq * math.Pow(2, (float64(note)-a4MidiNote)/numberOfNotes)
}
