package main

import (
	"fmt"
	"github.com/CarlKlagba/go-midi-synth/audio"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"log"
)

const sampleRate = 44100

type model struct {
	choices        []string
	cursor         int
	selected       int
	volume         float64
	volumeProgress progress.Model
}

func main() {
	err2 := audio.PlayNote()
	if err2 != nil {
		fmt.Printf("Error starting audio: %v", err2)
		log.Fatal(err2)
	}
	select {}
	/*p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}*/
}

type audioProcessOkMsg int
type audioProcessErrMsg struct{ error }

type volumeUpMsg float64
type volumeDownMsg float64

func (m model) Init() tea.Cmd {
	return playNote()
}

func playNote() func() tea.Msg {
	return func() tea.Msg {
		return audioProcessOkMsg(1)
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case audioProcessOkMsg:
		return m, nil

	case audioProcessErrMsg:
		log.Println("Arrêt du flux audio")
		return m, tea.Quit

	case volumeUpMsg:
		m.volume += 0.1

	case volumeDownMsg:
		m.volume -= 0.1

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		case "enter", " ":
			m.selected = m.cursor

		case "+", "right":
			return m, volumeUpCmd()

		case "-", "left":
			return m, volumeDownCmd()

		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	return m, nil
}

func volumeUpCmd() tea.Cmd {
	return func() tea.Msg {
		return volumeUpMsg(0.1)
	}
}

func volumeDownCmd() tea.Cmd {
	return func() tea.Msg {
		return volumeDownMsg(0.1)
	}
}

func initialModel() model {
	return model{
		choices:        []string{"Sine Wave", "Square Wave", "Triangle Wave", "Sawtooth Wave"},
		cursor:         0,
		selected:       0,
		volume:         0.5,
		volumeProgress: progress.New(progress.WithScaledGradient("#0d2f02", "#2ca506")),
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
	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		style := listStyle
		if m.selected == i {
			style = selectedStyle
		}

		line := cursorStyle.Render(cursor) + style.Render(" "+choice)

		s += line + "\n"
	}

	s += "\n" + m.volumeProgress.ViewAs(m.volume) + "\n"
	s += faintStyle.Render("\nPress space to select, q to quit.\n")

	return s
}
