package main

import (
	"github.com/gordonklaus/portaudio"
	"math"
	"time"
)

func main() {
	const sampleRate = 44100
	const freq = 440.0
	const amplitude = 0.3
	const duration = 30 // secondes

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
