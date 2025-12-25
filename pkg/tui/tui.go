package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"nyct-feed/pkg/gtfs"
	"nyct-feed/pkg/tui/splash"
	"nyct-feed/pkg/tui/stationlist"
)

type model struct {
	schedule    gtfs.Schedule
	stations    []gtfs.Stop
	stationList stationlist.Model
	loading     bool

	width  int
	height int
}

func NewModel() model {
	return model{loading: true}
}

func (m *model) Init() tea.Cmd {
	return getSchedule()
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case gotScheduleMsg:
		m.schedule = gtfs.Schedule(msg)
		m.stations = m.schedule.GetStations()
		m.stationList = stationlist.NewModel(m.stations, m.schedule.Routes)
		m.stationList.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
		m.loading = false
	}

	if m.loading == false {
		updatedModel, cmd := m.stationList.Update(msg)
		m.stationList = *updatedModel.(*stationlist.Model)
		return m, cmd
	}

	return m, nil
}

func (m *model) View() string {
	if m.loading {
		return lipgloss.NewStyle().
			Width(m.width).
			Height(m.height).
			Align(lipgloss.Center, lipgloss.Center).
			Render(splash.Model{}.View())
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, m.stationList.View())
}

type gotScheduleMsg gtfs.Schedule

func getSchedule() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(500 * time.Millisecond)
		schedule := gtfs.GetSchedule()
		return gotScheduleMsg(schedule)
	}
}
