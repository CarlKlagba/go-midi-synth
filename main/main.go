package main

import (
	"fmt"
	"github.com/gordonklaus/portaudio"
	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/reader"
	"gitlab.com/gomidi/rtmididrv"
	"log"
	"math"
	"sync/atomic"
	"time"
)

const sampleRate = 44100
const amplitude = 0.5
const noteOff = 0

const midiNoteOn byte = 0x90
const midiNoteOff byte = 0x80

type NotesPlayed struct {
	notes []uint8
}

var (
	lastPlayedNote    atomic.Uint32
	atomicPlayedNotes atomic.Value
	playedNotes       []uint8
	noteToFreq        = make(map[uint32]float64)
)

func midiNoteToFreq(note uint8) float64 {
	return 440.0 * math.Pow(2, (float64(note)-69)/12)
}

func main() {
	must(portaudio.Initialize())
	defer portaudio.Terminate()

	atomicPlayedNotes.Store(NotesPlayed{notes: make([]uint8, 0, 10)})

	// Initialiser les fréquences des notes MIDI
	for i := 0; i <= 127; i++ {
		noteToFreq[uint32(i)] = midiNoteToFreq(uint8(i))
	}
	noteToFreq[noteOff] = 0.0

	phase := 0.0
	updateDatedPhase := 0.0 // pour debug
	phaseStep := 1.0 / sampleRate
	var fadeInSample uint16 = 441 // 10ms de fade-in
	var fadeInCount uint16 = 0
	amp := amplitude
	stream, err := portaudio.OpenDefaultStream(0, 1, sampleRate, 256, func(out []float32) {
		notesPlayed := atomicPlayedNotes.Load().(NotesPlayed)

		for i := range out {
			if fadeInCount < fadeInSample {
				amp = amplitude * float64(fadeInCount) / float64(fadeInSample)
				fadeInCount++
			}

			o := float32(0.0)
			for _, v := range notesPlayed.notes {
				o += float32(amp * math.Sin(2*math.Pi*noteToFreq[uint32(v)]*phase))
			}

			out[i] = o

			_, updateDatedPhase = math.Modf(phase + phaseStep)

			if math.Abs(updateDatedPhase-phase) > 0.1 {
				phaseStep = -1 * phaseStep // Inversion du sens de la phase pour atténuer le clic
				_, updateDatedPhase = math.Modf(phase + phaseStep)
			}

			phase = updateDatedPhase

			//fmt.Println(phase)
		}
	})
	must(err)
	defer stream.Close()
	must(stream.Start())
	defer stream.Stop()

	drv, err := rtmididrv.New()
	if err != nil {
		log.Fatal(err)
	}
	defer must(drv.Close())

	ins, err := drv.Ins()
	if err != nil {
		log.Fatal(err)
	}

	if len(ins) == 0 {
		log.Fatal("Aucun port MIDI d'entrée trouvé")
	}

	in := ins[0]
	must(in.Open())

	defer func(in midi.In) {
		must(in.Close())
	}(in)

	rd := reader.New(
		reader.NoLogger(),
		reader.Each(func(pos *reader.Position, msg midi.Message) {
			midiBytes := msg.Raw()
			if len(midiBytes) < 3 {
				return
			}
			canal := midiBytes[0] & 0xF0
			note := midiBytes[1]
			velocity := midiBytes[2]
			fmt.Printf("Canal: 0x%X, Note: %d, Velocity: %d\n", canal, note, velocity)

			if canal == midiNoteOn && velocity > 0 { // Note ON
				playedNotes = append(playedNotes, note)
				n := NotesPlayed{notes: playedNotes}
				atomicPlayedNotes.Store(n)
			} else if (canal == midiNoteOff) || (canal == midiNoteOn && velocity == 0) { // Note OFF
				for i, n := range playedNotes {
					if n == note {
						playedNotes = append(playedNotes[:i], playedNotes[i+1:]...)
						break
					}
				}
				n := NotesPlayed{notes: playedNotes}
				atomicPlayedNotes.Store(n)

			}
		}),
	)

	fmt.Println("En attente de messages MIDI...")
	err1 := rd.ListenTo(in)
	must(err1)

	time.Sleep(1000 * time.Second)

}

/*
func main() {
	portaudio.Initialize()
	defer portaudio.Terminate()
	s := newStereoSine(256, 320, sampleRate)
	defer s.Close()
	chk(s.Start())
	time.Sleep(10 * time.Second)
	chk(s.Stop())
}

type stereoSine struct {
	*portaudio.Stream
	stepL, phaseL float64
	stepR, phaseR float64
}

func newStereoSine(freqL, freqR, sampleRate float64) *stereoSine {
	s := &stereoSine{nil, freqL / sampleRate, 0, freqR / sampleRate, 0}
	var err error
	s.Stream, err = portaudio.OpenDefaultStream(0, 2, sampleRate, 0, s.processAudio)
	chk(err)
	return s
}

func (g *stereoSine) processAudio(out [][]float32) {
	for i := range out[0] {
		out[0][i] = float32(math.Sin(2 * math.Pi * g.phaseL))
		_, g.phaseL = math.Modf(g.phaseL + g.stepL)
		out[1][i] = float32(math.Sin(2 * math.Pi * g.phaseR))
		_, g.phaseR = math.Modf(g.phaseR + g.stepR)
	}
}


func chk(err error) {
	if err != nil {
		panic(err)
	}
}
*/

/*
func main() {
	const sampleRate = 44100
	const freq = 440.0
	const amplitude = 0.3
	const duration = 3 // secondes

	portaudio.Initialize()
	defer portaudio.Terminate()

	stream, err := portaudio.OpenDefaultStream(0, 1, sampleRate, 0, func(out []float32) {
		for i := range out {
			t := float64(i+sampleRate*int(time.Now().UnixNano()/1e9)) / sampleRate
			out[i] = float32(amplitude * math.Sin(2*math.Pi*freq*t))
		}
	})
	if err != nil {
		panic(err)
	}
	defer stream.Close()

	stream.Start()
	time.Sleep(time.Second * duration)
	stream.Stop()
}

*/

func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}
