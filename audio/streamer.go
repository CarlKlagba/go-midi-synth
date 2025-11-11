package audio

import (
	"github.com/gordonklaus/portaudio"
	"log"
)

func StreamAudio(wp *WaveProcessor) error {
	must(portaudio.Initialize())
	defer portaudio.Terminate()

	stream, err := portaudio.OpenDefaultStream(
		0,
		1,
		sampleRate,
		256,
		wp.ProcessAudio,
	)

	if err != nil {
		return err
	}
	defer stream.Close()

	log.Println("Starting audio stream...")
	err = stream.Start()
	if err != nil {
		return err
	}

	log.Println("Audio stream started.")
	defer stream.Stop()
	return err
}
