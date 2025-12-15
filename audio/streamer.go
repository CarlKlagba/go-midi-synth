package audio

import (
	"github.com/gordonklaus/portaudio"
	"log"
)

func StreamAudio(wp *WaveProcessor) (*portaudio.Stream, error) {
	err := portaudio.Initialize()
	if err != nil {
		return nil, err
	}
	stream, err := portaudio.OpenDefaultStream(
		0,
		1,
		sampleRate,
		256,
		wp.ProcessAudio,
	)

	if err != nil {
		return nil, err
	}

	log.Println("Starting audio stream...")
	err = stream.Start()
	if err != nil {
		return nil, err
	}

	log.Println("Audio stream started.")
	return stream, nil
}

func CloseAudioStream(stream *portaudio.Stream) {
	log.Println("Closing audio stream...")
	_ = portaudio.Terminate()
	_ = stream.Close()
	_ = stream.Stop()
}
