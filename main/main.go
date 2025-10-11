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
	atomicPlayedNotes atomic.Value
	playedNotes       []uint8
	noteToFreq        = make(map[uint32]float64)
	noteToPhase       = make(map[uint32]float64)
	noteToFadeIn      = make(map[uint32]uint16)
)

func main() {
	must(portaudio.Initialize())
	defer portaudio.Terminate()

	atomicPlayedNotes.Store(NotesPlayed{notes: make([]uint8, 0, 10)})

	// Initialiser les fréquences des notes MIDI
	for i := 0; i <= 127; i++ {
		noteToFreq[uint32(i)] = midiNoteToFreq(uint8(i))
	}
	noteToFreq[noteOff] = 0.0

	// Initialiser les phases des notes MIDI
	for i := 0; i <= 127; i++ {
		noteToPhase[uint32(i)] = 0.0
	}

	// Initialiser les fades-in des notes MIDI
	for i := 0; i <= 127; i++ {
		noteToFadeIn[uint32(i)] = 0
	}

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
				freq := noteToFreq[uint32(v)]
				phase := noteToPhase[uint32(v)]
				step := freq / sampleRate

				o += float32(amp * math.Sin(2*math.Pi*phase))

				_, noteToPhase[uint32(v)] = math.Modf(phase + step)
			}

			out[i] = o
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

func midiNoteToFreq(note uint8) float64 {
	return 440.0 * math.Pow(2, (float64(note)-69)/12)
}

func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}
