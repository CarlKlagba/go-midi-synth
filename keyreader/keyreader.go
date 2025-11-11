package keyreader

import (
	"fmt"
	"github.com/CarlKlagba/go-midi-synth/audio"
	"github.com/eiannone/keyboard"
	"log"
	"os"
	"sync/atomic"
)

func Controls(atomicWaveform *atomic.Value) {
	go func() {
		err := keyboard.Open()
		if err != nil {
			log.Fatal(err)
		}
		defer keyboard.Close()
		fmt.Print("Press 's' for Sine wave")
		fmt.Print(", 'q' for Square wave")
		fmt.Print(", 't' for Triangle wave")
		fmt.Print(", 'w' for Sawtooth wave")
		fmt.Println(" or ESC or Ctrl+C to quit")
		for {
			r, key, err := keyboard.GetKey()
			if err != nil {
				log.Fatal(err)
			}
			switch key {
			case keyboard.KeyEsc:
				fmt.Println("Exit program")
				os.Exit(0)
			case keyboard.KeyCtrlC:
				fmt.Println("Exit program")
				os.Exit(0)
			default:
				switch r {
				case 's':
					fmt.Println("Waveform: Sine")
					atomicWaveform.Store(audio.Sine)
				case 'q':
					fmt.Println("Waveform: Square")
					atomicWaveform.Store(audio.Square)
				case 't':
					fmt.Println("Waveform: Triangle")
					atomicWaveform.Store(audio.Triangle)
				case 'w':
					fmt.Println("Waveform: Sawtooth")
					atomicWaveform.Store(audio.Sawtooth)
				}
			}
		}
	}()
}
