package audio

import (
	"math"
	"sync/atomic"
)

const sampleRate = 44100
const initialAmplitude = 0.8
const initialAttackCount = 4410 //10ms
const initialDecayCount = 4410  //10ms
const initialSustain = 0.3
const initialReleaseCount = 4410 //10ms
const maxVelocity = 127
const noteOff = 0

type NotesPlayed []MidiNote

type Waveform int8

const (
	Sine Waveform = iota
	Square
	Sawtooth
	Triangle
)

type WaveProcessor struct {
	AtomicPlayedNotes  *atomic.Pointer[[]MidiNote]
	AtomicWaveform     *atomic.Value
	atomicMaxAmp       *atomic.Value
	atomicAttackCount  *atomic.Uint32
	atomicDecayCount   *atomic.Uint32
	atomicSustain      *atomic.Value
	atomicReleaseCount *atomic.Uint32
	noteToFreq         map[uint8]float64
	noteToPhase        map[uint8]float64
	velocityToAmp      map[uint8]float64
	noteToAttack       map[uint8]uint32
	noteToDecay        map[uint8]uint32
	noteToRelease      map[uint8]uint32
}

func NewWaveProcessor() *WaveProcessor {
	noteToFreq := make(map[uint8]float64, 128)
	noteToPhase := make(map[uint8]float64, 128)
	velocityToAmp := make(map[uint8]float64, 128)
	noteToAttack := make(map[uint8]uint32, 128)
	noteToDecay := make(map[uint8]uint32, 128)
	noteToRelease := make(map[uint8]uint32, 128)

	notes := make([]MidiNote, 0, 10) //careful we only allocate 10, so we might only be able to play 10 notes at the time
	var atomicPlayedNotes atomic.Pointer[[]MidiNote]
	atomicPlayedNotes.Store(&notes)

	atomicWaveform := &atomic.Value{}
	atomicWaveform.Store(Sine)
	atomicMaxAmp := &atomic.Value{}
	atomicMaxAmp.Store(initialAmplitude)
	atomicAttackCount := &atomic.Uint32{}
	atomicAttackCount.Store(initialAttackCount)
	atomicDecayCount := &atomic.Uint32{}
	atomicDecayCount.Store(initialDecayCount)
	atomicSustain := &atomic.Value{}
	atomicSustain.Store(initialSustain)
	atomicReleaseCount := &atomic.Uint32{}
	atomicReleaseCount.Store(initialReleaseCount)

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
		noteToDecay[uint8(i)] = initialDecayCount
	}

	for i := 0; i <= 127; i++ {
		noteToRelease[uint8(i)] = initialReleaseCount
	}

	return &WaveProcessor{
		AtomicPlayedNotes:  &atomicPlayedNotes,
		AtomicWaveform:     atomicWaveform,
		atomicMaxAmp:       atomicMaxAmp,
		atomicAttackCount:  atomicAttackCount,
		atomicDecayCount:   atomicDecayCount,
		atomicSustain:      atomicSustain,
		atomicReleaseCount: atomicReleaseCount,
		noteToFreq:         noteToFreq,
		noteToPhase:        noteToPhase,
		velocityToAmp:      velocityToAmp,
		noteToAttack:       noteToAttack,
		noteToDecay:        noteToDecay,
		noteToRelease:      noteToRelease,
	}
}

func (w *WaveProcessor) ProcessAudio(out []float32) {
	notesPlayed := *w.AtomicPlayedNotes.Load() // TODO see how comfortable we are about the pointer
	waveform := w.AtomicWaveform.Load().(Waveform)
	maxAmp := w.atomicMaxAmp.Load().(float64)
	attackCount := w.atomicAttackCount.Load()
	decayCount := w.atomicDecayCount.Load()
	sustain := w.atomicSustain.Load().(float64)
	releaseCount := w.atomicReleaseCount.Load()

	for i := range out {
		o := float32(0.0)
		for _, midiNote := range notesPlayed {
			if !midiNote.On && w.noteToRelease[midiNote.Note] == 0 {
				//Do not remove note when off and release over to avoid race conditions with midiReader. Keep all atomic read only
				continue
			}
			if midiNote.On && w.noteToRelease[midiNote.Note] != releaseCount {
				w.noteToRelease[midiNote.Note] = releaseCount
			}
			if !midiNote.On && w.noteToAttack[midiNote.Note] != 0 {
				w.noteToAttack[midiNote.Note] = 0
			}
			if !midiNote.On && w.noteToDecay[midiNote.Note] != decayCount {
				w.noteToDecay[midiNote.Note] = decayCount
			}

			freq := w.noteToFreq[midiNote.Note]
			phase := w.noteToPhase[midiNote.Note]
			amp := w.velocityToAmp[midiNote.Velocity] * maxAmp

			if midiNote.On && w.noteToAttack[midiNote.Note] < attackCount {
				amp = amp * (float64(w.noteToAttack[midiNote.Note]) / float64(attackCount))
				w.noteToAttack[midiNote.Note]++
			}

			if midiNote.On &&
				w.noteToAttack[midiNote.Note] >= attackCount && // The attack phase is done
				w.noteToDecay[midiNote.Note] > 0 {
				decay := (1.0 - sustain) * float64(w.noteToDecay[midiNote.Note]) / float64(decayCount)
				amp = amp * (decay + sustain)
				if w.noteToDecay[midiNote.Note] > 0 {
					w.noteToDecay[midiNote.Note]--
				}
			}

			if !midiNote.On && w.noteToRelease[midiNote.Note] > 0 {
				amp = amp * sustain * (float64(w.noteToRelease[midiNote.Note]) / float64(releaseCount))
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

func (w *WaveProcessor) SetVolume(volume float64) {
	if volume < 0.0 {
		volume = 0.0
	}
	if volume >= 1.0 {
		volume = 1.0
	}
	w.atomicMaxAmp.Store(volume)
}

func (w *WaveProcessor) GetVolume() float64 {
	return w.atomicMaxAmp.Load().(float64)
}

func (w *WaveProcessor) SetAttackTime(millitsec float64) {
	if millitsec < 5.0 {
		millitsec = 5.0
	}
	if millitsec >= 100.0 {
		millitsec = 100.0
	}
	val := uint32(millitsec * 440)
	w.atomicAttackCount.Store(val)
}

func (w *WaveProcessor) GetAttackTime() float64 {
	r := w.atomicAttackCount.Load()
	return float64(r) / 440.0
}

func (w *WaveProcessor) GetDecayTime() float64 {
	r := w.atomicDecayCount.Load()
	return float64(r) / 440.0
}

func (w *WaveProcessor) SetDecayTime(millitsec float64) {
	if millitsec < 5.0 {
		millitsec = 5.0
	}
	if millitsec >= 100.0 {
		millitsec = 100.0
	}
	val := uint32(millitsec * 440)
	w.atomicDecayCount.Store(val)
}

func (w *WaveProcessor) GetSustain() float64 {
	r := w.atomicSustain.Load().(float64)
	return float64(r) / 440.0
}

func (w *WaveProcessor) SetSustain(sustain float64) {
	if sustain < 0.0 {
		sustain = 0.0
	}
	if sustain >= 1.0 {
		sustain = 1.0
	}
	w.atomicSustain.Store(sustain)
}

func (w *WaveProcessor) SetReleaseTime(millitsec float64) {
	if millitsec <= 5.0 {
		millitsec = 5.0
	}
	if millitsec >= 100.0 {
		millitsec = 100.0
	}
	val := uint32(millitsec * 440)
	w.atomicReleaseCount.Store(val)
}

func (w *WaveProcessor) GetReleaseTime() float64 {
	r := w.atomicReleaseCount.Load()
	return float64(r) / 440.0
}
