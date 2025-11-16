package main

import (
	"fmt"
	"github.com/CarlKlagba/go-midi-synth/audio"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"log"
	"os"
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
}

func main() {

	wp := audio.NewWaveProcessor()
	stream, err := audio.StreamAudio(wp)
	if err != nil {
		log.Fatalf("Failed to start audio stream: %v", err)
	}
	defer audio.CloseAudioStream(stream)

	midiNotesChan := make(chan []uint8, 30)
	defer close(midiNotesChan)
	err = audio.StartReadingMidiMessages(wp, midiNotesChan)
	if err != nil {
		log.Fatalf("Failed to start midi reading: %v", err)
	}
	defer audio.CloseMidiReader()

	p := tea.NewProgram(initialModel(wp, midiNotesChan))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error starting the TUI: %v", err)
		os.Exit(1)
	}
}

func initialModel(wp *audio.WaveProcessor, notesReceiver <-chan []uint8) model {
	return model{
		waves:          []audio.Waveform{audio.Sine, audio.Square, audio.Triangle, audio.Sawtooth},
		cursor:         0,
		selected:       0,
		notesReceiver:  notesReceiver,
		notesPlayed:    make([]uint8, 0),
		volume:         0.5,
		volumeProgress: progress.New(progress.WithScaledGradient("#0d2f02", "#2ca506")),
		waveProcessor:  wp,
	}
}

func (m model) Init() tea.Cmd {
	return waitForNoteCmd(&m)
}

func waitForNoteCmd(m *model) tea.Cmd {
	return func() tea.Msg {
		np := <-m.notesReceiver
		return playNotesMsg(np)
	}
}

type volumeSetAtMsg float64
type waveformSelectedMsg int
type playNotesMsg []uint8

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case volumeSetAtMsg:
		m.volume = float64(msg)

	case waveformSelectedMsg:
		m.selected = int(msg)

	case playNotesMsg:
		m.notesPlayed = msg
		return m, waitForNoteCmd(&m)

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.waves)-1 {
				m.cursor++
			}

		case "enter", " ":
			return m, selectWaveformCmd(&m)

		case "+", "right":
			return m, volumeUpCmd(&m)

		case "-", "left":
			return m, volumeDownCmd(&m)

		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	return m, nil
}

func volumeUpCmd(m *model) tea.Cmd {
	m.waveProcessor.SetVolume(m.volume + 0.01)
	updatedVolume := m.waveProcessor.GetVolume()
	return func() tea.Msg {
		return volumeSetAtMsg(updatedVolume)
	}
}

func volumeDownCmd(m *model) tea.Cmd {
	m.waveProcessor.SetVolume(m.volume - 0.01)
	updatedVolume := m.waveProcessor.GetVolume()
	return func() tea.Msg {
		return volumeSetAtMsg(updatedVolume)
	}
}

func selectWaveformCmd(m *model) tea.Cmd {
	m.waveProcessor.AtomicWaveform.Store(m.waves[m.cursor])
	return func() tea.Msg {
		return waveformSelectedMsg(m.cursor)
	}
}

var (
	headerStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("232")).Background(lipgloss.Color("34")).Padding(0, 2, 0, 2).Bold(true)
	cursorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("34"))
	listStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("28"))
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("34")).Bold(true)
	faintStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("34")).Faint(true)
)

func (m model) View() string {
	s := headerStyle.Render("Sexy Synth") + "\n"
	for i, wave := range m.waves {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		style := listStyle
		if m.selected == i {
			style = selectedStyle
		}

		line := cursorStyle.Render(cursor) + style.Render(" "+toString(wave))

		s += line + "\n"
	}

	s += "\n" + m.volumeProgress.ViewAs(m.volume) + "\n"
	s += "\n \t" + fmt.Sprintln(m.notesPlayed) + "\n"
	s += faintStyle.Render("\nPress space to select, q to quit.\n")

	return s
}

func toString(wave audio.Waveform) string {
	switch wave {
	case audio.Sine:
		return "Sine Wave"
	case audio.Square:
		return "Square Wave"
	case audio.Triangle:
		return "Triangle Wave"
	case audio.Sawtooth:
		return "Sawtooth Wave"
	default:
		return "Unknown Waveform"
	}
}
