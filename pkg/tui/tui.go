package tui

import (
	"nyct-feed/pkg/gtfs"
	splash "nyct-feed/pkg/tui/splash"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	docStyle lipgloss.Style
	schedule gtfs.Schedule
	stations []gtfs.Stop
	loading  bool
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
		m.schedule = gtfs.Schedule(msg)
		m.stations = m.schedule.GetStations()
		m.loading = false
	}

	var cmd tea.Cmd
	return m, cmd
}

func (m model) View() string {
	if m.loading {
	return m.docStyle.
		Align(lipgloss.Center, lipgloss.Center).
		Render(splash.Model{}.View())
	}
	return "loaded!"
}

type gotScheduleMsg gtfs.Schedule

func getSchedule() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(200 * time.Millisecond)
		schedule := gtfs.GetSchedule()
		return gotScheduleMsg(schedule)
	}
}
