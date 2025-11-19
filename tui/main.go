package main

import (
	"fmt"
	"github.com/CarlKlagba/go-midi-synth/audio"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"log"
	"os"
	"strconv"
	"strings"
)

type progFlags struct {
	noNotesDisplay bool
}

var flags = progFlags{
	noNotesDisplay: false,
}

func (p *progFlags) flag(progArg []string) {
	for _, a := range progArg {
		if a == "--no-note-display" || a == "-nd" {
			p.noNotesDisplay = true
			log.Println("No Notes Display")
		}
	}
}

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
}

func main() {
	args := os.Args[1:]
	flags.flag(args)

	wp := audio.NewWaveProcessor()
	stream, err := audio.StreamAudio(wp)
	if err != nil {
		log.Fatalf("Failed to start audio stream: %v", err)
	}
	defer audio.CloseAudioStream(stream)

	var midiNotesChan chan []uint8 = nil
	if !flags.noNotesDisplay {
		//Putting the buffer size to 100 seems to fix the issue with clipping but need to look further into it
		midiNotesChan = make(chan []uint8, 100)
		defer close(midiNotesChan)
	}

	err = audio.StartReadingMidiMessages(wp, midiNotesChan)
	if err != nil {
		log.Fatalf("Failed to start midi reading: %v", err)
	}
	defer audio.CloseMidiReader()

	p := tea.NewProgram(initialModel(wp, midiNotesChan))
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error starting the TUI: %v", err)
	}
}

func initialModel(wp *audio.WaveProcessor, notesReceiver <-chan []uint8) model {
	return model{
		waves:          []audio.Waveform{audio.Sine, audio.Square, audio.Triangle, audio.Sawtooth},
		cursor:         0,
		selected:       0,
		notesReceiver:  notesReceiver,
		notesPlayed:    make([]uint8, 0),
		volume:         0.5, //Recuperer cette valeur du wave processor
		volumeProgress: progress.New(progress.WithScaledGradient("#0d2f02", "#2ca506")),
		waveProcessor:  wp,
		releaseTime:    wp.GetReleaseTime(),
	}
}

func (m model) Init() tea.Cmd {
	return waitForNoteCmd(&m)
}

func waitForNoteCmd(m *model) tea.Cmd {
	if m.notesReceiver != nil {
		return nil
	}
	return func() tea.Msg {
		np := <-m.notesReceiver
		return playNotesMsg(np)
	}
}

type playNotesMsg []uint8
type waveformSelectedMsg int
type volumeSetAtMsg float64
type releaseTimeSetAtMsg float64

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case volumeSetAtMsg:
		m.volume = float64(msg)

	case releaseTimeSetAtMsg:
		m.releaseTime = float64(msg)

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

		case "r":
			return m, releaseDownCmd(&m)
		case "R":
			return m, releaseUpCmd(&m)

		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	return m, nil
}

func selectWaveformCmd(m *model) tea.Cmd {
	m.waveProcessor.AtomicWaveform.Store(m.waves[m.cursor])
	return func() tea.Msg {
		return waveformSelectedMsg(m.cursor)
	}
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
func releaseUpCmd(m *model) tea.Cmd {
	m.waveProcessor.SetReleaseTime(m.releaseTime + 5.0)
	return func() tea.Msg {
		return releaseTimeSetAtMsg(m.waveProcessor.GetReleaseTime())
	}
}

func releaseDownCmd(m *model) tea.Cmd {
	m.waveProcessor.SetReleaseTime(m.releaseTime - 5.0)
	return func() tea.Msg {
		return releaseTimeSetAtMsg(m.waveProcessor.GetReleaseTime())
	}
}

var (
	headerStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("232")).Background(lipgloss.Color("34")).Padding(0, 2, 0, 2).Bold(true)
	cursorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("34"))
	basicTextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("34"))
	listStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("28"))
	boldTextStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("34")).Bold(true)
	notesStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("34")).Bold(true)
	faintStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("34")).Faint(true)
)

func (m model) View() string {
	var full strings.Builder

	full.WriteString(
		fmt.Sprintf("%s\t\t%s \n", headerStyle.Render("Sexy Synth"), notesStyle.Render(DisplayNotes(m.notesPlayed))))
	var waves strings.Builder
	for i, wave := range m.waves {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		style := listStyle
		if m.selected == i {
			style = boldTextStyle
		}

		waves.WriteString(cursorStyle.Render(cursor) + style.Render(" "+toString(wave)))
		waves.WriteString("\n")
	}

	release := lipgloss.JoinVertical(lipgloss.Top,
		boldTextStyle.Render("Release"),
		basicTextStyle.Render(fmt.Sprintf("%sms", strconv.FormatFloat(m.releaseTime, 'f', 1, 32))),
		faintStyle.Render("r - R"))

	full.WriteString(lipgloss.JoinHorizontal(lipgloss.Left, waves.String(), "       ", release))
	full.WriteString("\n" + m.volumeProgress.ViewAs(m.volume) + "\n")
	full.WriteString(faintStyle.Render("\nPress space to select, q to quit.\n"))

	return full.String()
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

var doremi = [12]string{
	"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B",
}

func DisplayNotes(notes []uint8) string {
	var sb strings.Builder
	for _, n := range notes {
		sb.WriteString(displayNote(n))
		sb.WriteString(" ")
	}
	return sb.String()
}

func displayNote(n uint8) string {
	u := int(n % 12)
	r := int(n/12) - 1
	var sb strings.Builder
	sb.WriteString(doremi[u])
	sb.WriteString(strconv.Itoa(r))
	return sb.String()
}
