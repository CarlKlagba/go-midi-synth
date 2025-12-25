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
	noteToFreq         [128]float32
	noteToPhase        [128]float32
	velocityToAmp      [128]float32
	noteToAttack       [128]uint16
	noteToDecay        [128]uint16
	noteToRelease      [128]uint16
}

func NewWaveProcessor() *WaveProcessor {
	return NewWaveProcessorWith(
		sampleRate,
		initialAmplitude,
		initialAttackCount,
		initialDecayCount,
		initialSustain,
		initialReleaseCount,
	)
}

func NewWaveProcessorWith(
	sampleRate uint32,
	initialAmplitude float32,
	initialAttackCount uint16,
	initialDecayCount uint16,
	initialSustain float32,
	initialReleaseCount uint16,
) *WaveProcessor {
	var noteToFreq [128]float32
	var noteToPhase [128]float32
	var velocityToAmp [128]float32
	var noteToAttack [128]uint16
	var noteToDecay [128]uint16
	var noteToRelease [128]uint16

	notes := make([]MidiNote, 0, 10) //careful we only allocate 10, so we might only be able to play 10 notes at the time
	var atomicPlayedNotes atomic.Pointer[[]MidiNote]
	atomicPlayedNotes.Store(&notes)

	atomicWaveform := &atomic.Value{}
	atomicWaveform.Store(Sine)
	atomicMaxAmp := &atomic.Value{}
	atomicMaxAmp.Store(initialAmplitude)
	atomicAttackCount := &atomic.Uint32{}
	atomicAttackCount.Store(uint32(initialAttackCount))
	atomicDecayCount := &atomic.Uint32{}
	atomicDecayCount.Store(uint32(initialDecayCount))
	atomicSustain := &atomic.Value{}
	atomicSustain.Store(initialSustain)
	atomicReleaseCount := &atomic.Uint32{}
	atomicReleaseCount.Store(uint32(initialReleaseCount))

	for i := 0; i <= 127; i++ {
		noteToFreq[i] = midiNoteToFreq(i)
	}
	noteToFreq[noteOff] = 0.0

	for i := 0; i <= 127; i++ {
		noteToPhase[i] = 0.0
	}

	for i := 0; i <= 127; i++ {
		velocityToAmp[i] = midiVelocityToAmplitude(i) * initialAmplitude
	}

	for i := 0; i <= 127; i++ {
		noteToAttack[i] = 0
	}

	for i := 0; i <= 127; i++ {
		noteToDecay[i] = initialDecayCount
	}

	for i := 0; i <= 127; i++ {
		noteToRelease[i] = initialReleaseCount
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
	w.processAudio(out, nil)
}

func (w *WaveProcessor) ProcessAudioTrackAmplitude(out []float32, ampOut []float32) {
	w.processAudio(out, ampOut)
}

func (w *WaveProcessor) processAudio(out []float32, ampOut []float32) {
	notesPlayed := *w.AtomicPlayedNotes.Load() //TODO see how comfortable we are about the pointer
	waveform := w.AtomicWaveform.Load().(Waveform)
	maxAmp := w.atomicMaxAmp.Load().(float32)
	attackCount := uint16(w.atomicAttackCount.Load())
	decayCount := uint16(w.atomicDecayCount.Load())
	sustain := w.atomicSustain.Load().(float32)
	releaseCount := uint16(w.atomicReleaseCount.Load())

	for i := range out {
		o := float32(0.0)
		for _, midiNote := range notesPlayed {
			if !midiNote.On && w.noteToRelease[midiNote.Note] == 0 {
				//Do not remove note when off and release over to avoid race conditions with midiReader. Keep all atomic read only
				continue
			}
			if midiNote.On && w.noteToRelease[midiNote.Note] != releaseCount { //Reinitialise the release when a note starts again
				w.noteToRelease[midiNote.Note] = releaseCount
			}
			if !midiNote.On && w.noteToAttack[midiNote.Note] != 0 { //Reinitialise the attack when the note is off
				w.noteToAttack[midiNote.Note] = 0
			}
			if !midiNote.On && w.noteToDecay[midiNote.Note] != decayCount {
				w.noteToDecay[midiNote.Note] = decayCount
			}

			freq := w.noteToFreq[midiNote.Note]
			phase := w.noteToPhase[midiNote.Note]
			amp := w.velocityToAmp[midiNote.Velocity] * maxAmp

			if midiNote.On && w.noteToAttack[midiNote.Note] < attackCount { // The Attack phase
				amp = amp * (float32(w.noteToAttack[midiNote.Note]) / float32(attackCount))
				w.noteToAttack[midiNote.Note]++
			}

			if midiNote.On && w.noteToAttack[midiNote.Note] >= attackCount { // The attack phase is done
				decay := (1.0 - sustain) * float32(w.noteToDecay[midiNote.Note]) / float32(decayCount)
				amp = amp * (decay + sustain)
				if w.noteToDecay[midiNote.Note] > 0 {
					w.noteToDecay[midiNote.Note]--
				}
			}

			if !midiNote.On && w.noteToRelease[midiNote.Note] > 0 { // The release phase starts
				amp = amp * sustain * (float32(w.noteToRelease[midiNote.Note]) / float32(releaseCount))
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
			_, p1 := math.Modf(float64(phase + step))
			w.noteToPhase[midiNote.Note] = float32(p1)

			if ampOut != nil { //should be a matrix but lets just handle one note for now
				ampOut[i] = amp
			}
		}
		out[i] = o
	}
}

func sinWave(phase float32) float32 {
	return float32(math.Sin(2 * math.Pi * float64(phase)))
}

func sqrWave(phase float32) float32 {
	return float32(math.Copysign(1, math.Sin(2*math.Pi*float64(phase))))
}

func triWave(phase float32) float32 {
	return float32(4*math.Abs(float64(phase)-0.5) - 1)
}

func sawWave(phase float32) float32 {
	return phase - 0.5
}

func midiVelocityToAmplitude(velocity int) float32 {
	return float32(float32(velocity) / float32(maxVelocity))
}

func midiNoteToFreq(note int) float32 {
	const a4Freq = 440.0
	const a4MidiNote = 69.0
	const numberOfNotes = 12.0
	return float32(a4Freq * math.Pow(2, (float64(note)-a4MidiNote)/numberOfNotes))
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
