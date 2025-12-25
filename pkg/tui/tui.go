package tui

import (
	"time"

	"nyct-feed/pkg/gtfs"
	splash "nyct-feed/pkg/tui/splash"
	stationlist "nyct-feed/pkg/tui/stationlist"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	docStyle    lipgloss.Style
	schedule    gtfs.Schedule
	stations    []gtfs.Stop
	stationList stationlist.Model
	loading     bool
}

func NewModel() model {
	return model{
		loading:  true,
		docStyle: lipgloss.NewStyle().Margin(1, 2),
	}
}

func (m model) Init() tea.Cmd {
	return getSchedule()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := m.docStyle.GetFrameSize()
		m.docStyle = m.docStyle.Width(msg.Width - h).Height(msg.Height - v)
	case gotScheduleMsg:
		listHeight := m.docStyle.GetHeight() - m.docStyle.GetVerticalMargins()
		m.schedule = gtfs.Schedule(msg)
		m.stations = m.schedule.GetStations()
		m.stationList = stationlist.NewModel(m.stations, m.schedule.Routes, listHeight)
		m.loading = false
	}

	if m.loading == false {
		updatedModel, cmd := m.stationList.Update(msg)
    m.stationList = updatedModel.(stationlist.Model)
		return m, cmd
	}

	return m, nil
}

func (m model) View() string {
	if m.loading {
		return m.docStyle.
			Align(lipgloss.Center, lipgloss.Center).
			Render(splash.Model{}.View())
	}
	return m.docStyle.Render(m.stationList.View())
}

type gotScheduleMsg gtfs.Schedule

func getSchedule() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(200 * time.Millisecond)
		schedule := gtfs.GetSchedule()
		return gotScheduleMsg(schedule)
	}
}
