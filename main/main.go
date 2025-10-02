package main

import (
	"fmt"
	"github.com/gordonklaus/portaudio"
	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/reader"
	"gitlab.com/gomidi/rtmididrv"
	"log"
	"math"
	"sync"
	"sync/atomic"
	"time"
)

const sampleRate = 44100
const amplitude = 0.5

var (
	notesMu        sync.Mutex
	lastPlayedNote atomic.Uint32
	noteToFreq     = make(map[uint32]float64)
)

func midiNoteToFreq(note uint8) float64 {
	return 440.0 * math.Pow(2, (float64(note)-69)/12)
}

func main() {
	must(portaudio.Initialize())
	defer portaudio.Terminate()

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

	// Initialiser les fréquences des notes MIDI
	for i := 0; i <= 127; i++ {
		noteToFreq[uint32(i)] = midiNoteToFreq(uint8(i))
	}

	phase := 0.0
	stream, err := portaudio.OpenDefaultStream(0, 1, sampleRate, 0, func(out []float32) {
		for i := range out {
			freq := noteToFreq[lastPlayedNote.Load()]
			out[i] = float32(amplitude * math.Sin(2*math.Pi*freq*phase))
			phase += 1.0 / sampleRate
			if phase >= 1.0 {
				phase -= 1.0
			}
		}
	})
	must(err)
	defer stream.Close()
	must(stream.Start())
	defer stream.Stop()

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
			if canal == 0x90 && velocity > 0 { // Note ON
				lastPlayedNote.Store(uint32(note))
			} else if (canal == 0x80) || (canal == 0x90 && velocity == 0) { // Note OFF
				if lastPlayedNote.Load() == uint32(note) {
					lastPlayedNote.Store(0)
				}
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
