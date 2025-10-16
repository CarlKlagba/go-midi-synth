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
		fmt.Println("Appuie sur 's' pour sinusoïdale, 'q' pour carrée, 'ESC' pour quitter")
		for {
			r, key, err := keyboard.GetKey()
			if err != nil {
				log.Fatal(err)
			}
			switch key {
			case keyboard.KeyEsc:
				fmt.Println("Arrêt du programme")
				os.Exit(0)
			case keyboard.KeyCtrlC:
				fmt.Println("Arrêt du programme")
				os.Exit(0)
			default:
				switch r {
				case 's':
					fmt.Println("Forme d'onde: sinusoïdale")
					atomicWaveform.Store(audio.Sine)
				case 'q':
					fmt.Println("Forme d'onde: carrée")
					atomicWaveform.Store(audio.Square)
				case 't':
					fmt.Println("Forme d'onde: Triangular")
					atomicWaveform.Store(audio.Triangle)
				case 'w':
					fmt.Println("Forme d'onde: Sawtooth")
					atomicWaveform.Store(audio.Sawtooth)
				}
			}
		}
	}()
}
