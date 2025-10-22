package main

import (
	"fmt"
	"github.com/eiannone/keyboard"
	"log"
	"os"
	"time"
)

func main() {
	fmt.Println("Démarrage du programme")

	controls()

	time.Sleep(20 * time.Minute)
}
func controls() {
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
			default:
				switch r {
				case 's':
					//wp.setWaveform("sine")
					fmt.Println("Forme d'onde: sinusoïdale")
				case 'q':
					//wp.setWaveform("square")
					fmt.Println("Forme d'onde: carrée")
				}
			}
		}
	}()

}
