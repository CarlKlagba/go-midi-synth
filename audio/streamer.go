package audio

import "github.com/gordonklaus/portaudio"

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
	must(err)
	defer stream.Close()
	must(stream.Start())
	defer stream.Stop()
	return err
}
