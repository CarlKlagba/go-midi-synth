package audio

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"testing"
)

func test_ProcessesAudio_no_notes_no_out(t *testing.T) {
	wp := NewWaveProcessor()

	out := make([]float32, 256) // Is this the right size?
	wp.ProcessAudio(out)

	expected := make([]float32, 256)
	for i := range expected {
		expected[i] = 0
	}

	if !equalFloat32Slices(expected, out) {
		t.Errorf("Expected %v, got %v", expected, out)
	}
}

func equalFloat32Slices(expected []float32, out []float32) bool {
	for i := range expected {
		if expected[i] != out[i] {
			return false
		}
	}
	return true
}

func Test_ProcessAudio_With_a_Note(t *testing.T) {
	wp := NewWaveProcessor()

	notes := make([]MidiNote, 1)
	notes[0] = MidiNote{Note: 69, Velocity: 50, On: true}

	wp.AtomicPlayedNotes.Store(&notes)

	out := make([]float32, 256)    // Is this the right size?
	ampOut := make([]float32, 256) // Is this the right size?

	if len(ampOut) == 0 {
		fmt.Println("[]")
		return
	}

	fo, err := os.Create("waveprocess_test_data/waveprocess_amp.csv")
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()

	w := bufio.NewWriter(fo)

	w.WriteString("x,f\n")

	n := 0
	for range 440 {

		wp.ProcessAudioTrackAmplitude(out, ampOut)
		i := 0
		for i < len(ampOut) {
			tmp := fmt.Sprintf("%d,%f\n", i+len(ampOut)*n, ampOut[i])
			_, err := w.WriteString(tmp)
			if err != nil {
				panic("fail writing in the file")
			}

			i = i + 10
		}
		n++
	}
	w.Flush()

	cmd := exec.Command("sh", "display_wave.sh")
	cmd.Dir = "waveprocess_test_data"
	err1 := cmd.Run()

	if err1 != nil {
		formattedErr := fmt.Sprintf("failed to run command '%s' with args %v in directory '%s': %v", cmd.Path, cmd.Args, cmd.Dir, err1)
		fmt.Fprintln(os.Stderr, formattedErr)
		// Ensure non-zero exit in case log.Fatalf above is changed in the future.
		os.Exit(1)
	}
}
