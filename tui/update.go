package tui

import tea "github.com/charmbracelet/bubbletea"

type playNotesMsg []uint8
type waveformSelectedMsg int
type volumeSetAtMsg float64
type releaseTimeSetAtMsg float64
type attackTimeSetAtMsg float64

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case volumeSetAtMsg:
		m.volume = float64(msg)

	case releaseTimeSetAtMsg:
		m.releaseTime = float64(msg)

	case attackTimeSetAtMsg:
		m.attackTime = float64(msg)

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

		case "a":
			return m, attackDownCmd(&m)
		case "A":
			return m, attackUpCmd(&m)

		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	return m, nil
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

func attackUpCmd(m *model) tea.Cmd {
	m.waveProcessor.SetAttackTime(m.attackTime + 5.0)
	return func() tea.Msg {
		return attackTimeSetAtMsg(m.waveProcessor.GetAttackTime())
	}
}

func attackDownCmd(m *model) tea.Cmd {
	m.waveProcessor.SetAttackTime(m.attackTime - 5.0)
	return func() tea.Msg {
		return attackTimeSetAtMsg(m.waveProcessor.GetAttackTime())
	}
}
