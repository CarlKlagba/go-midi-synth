package tui

import (
	"github.com/CarlKlagba/go-midi-synth/audio"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	waves          []audio.Waveform
	cursor         int
	selected       int
	notesReceiver  <-chan []uint8
	notesPlayed    []uint8
	volume         float64
	volumeProgress progress.Model
	waveProcessor  *audio.WaveProcessor
	releaseTime    float64
	attackTime     float64
}

func InitialModel(wp *audio.WaveProcessor, notesReceiver <-chan []uint8) tea.Model {
	return model{
		waves:          []audio.Waveform{audio.Sine, audio.Square, audio.Triangle, audio.Sawtooth},
		cursor:         0,
		selected:       0,
		notesReceiver:  notesReceiver,
		notesPlayed:    make([]uint8, 0),
		volume:         wp.GetVolume(),
		volumeProgress: progress.New(progress.WithScaledGradient("#0d2f02", "#2ca506")),
		waveProcessor:  wp,
		releaseTime:    wp.GetReleaseTime(),
		attackTime:     wp.GetAttackTime(),
	}
}

func (m model) Init() tea.Cmd {
	return waitForNoteCmd(&m)
}
