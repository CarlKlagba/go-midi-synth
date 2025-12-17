package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/CarlKlagba/go-midi-synth/audio"
	"github.com/charmbracelet/lipgloss"
)

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

	attack := lipgloss.JoinVertical(lipgloss.Top,
		boldTextStyle.Render("Attack"),
		basicTextStyle.Render(fmt.Sprintf("%sms", strconv.FormatFloat(m.attackTime, 'f', 1, 32))),
		faintStyle.Render("a - A"))
	decay := lipgloss.JoinVertical(lipgloss.Top,
		boldTextStyle.Render("Decay"),
		boldTextStyle.Render(fmt.Sprintf("%sms", strconv.FormatFloat(m.decayTime, 'f', 1, 32))),
		faintStyle.Render("d - D"))
	sustain := lipgloss.JoinVertical(lipgloss.Top,
		boldTextStyle.Render("Sustain"),
		boldTextStyle.Render(fmt.Sprintf("%s%%", strconv.FormatFloat(m.sustain, 'f', 1, 32))),
		faintStyle.Render("s - S"))
	release := lipgloss.JoinVertical(lipgloss.Top,
		boldTextStyle.Render("Release"),
		basicTextStyle.Render(fmt.Sprintf("%sms", strconv.FormatFloat(m.releaseTime, 'f', 1, 32))),
		faintStyle.Render("r - R"))

	adsrSection := lipgloss.JoinHorizontal(lipgloss.Top, attack, "     ", decay, "     ", sustain, "     ", release)

	full.WriteString(lipgloss.JoinHorizontal(lipgloss.Left, waves.String(), "       ", adsrSection))
	full.WriteString("\n-" + m.volumeProgress.ViewAs(m.volume) + "+\n")
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
		sb.WriteString(DisplayNote(n))
		sb.WriteString(" ")
	}
	return sb.String()
}

func DisplayNote(n uint8) string {
	u := int(n % 12)
	r := int(n/12) - 1
	var sb strings.Builder
	sb.WriteString(doremi[u])
	sb.WriteString(strconv.Itoa(r))
	return sb.String()
}
