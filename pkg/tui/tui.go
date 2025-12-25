package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"nyct-feed/pkg/gtfs"
	"nyct-feed/pkg/tui/departuretable"
	"nyct-feed/pkg/tui/splash"
	"nyct-feed/pkg/tui/stationlist"
)

type model struct {
	schedule        gtfs.Schedule
	scheduleLoading bool
	stations        []gtfs.Stop
	selectedStation gtfs.Stop
	realtime        []gtfs.RealtimeFeed
	realtimeLoading bool
	departures      []gtfs.Departure
	stationList     stationlist.Model
	departureTable  departuretable.Model

	width  int
	height int
}

func NewModel() model {
	return model{
		scheduleLoading: true,
		realtimeLoading: true,
	}
}

func (m *model) Init() tea.Cmd {
	return tea.Batch(getSchedule(), getRealtime())
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
		m.scheduleLoading = false
	case gotRealtimeMsg:
		m.realtime = []gtfs.RealtimeFeed(msg)
		m.departures = gtfs.FindDepartures([]string{"635N"}, m.realtime)
		m.departureTable = departuretable.NewModel(m.departures)
		m.departureTable.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
		m.realtimeLoading = false
	case stationlist.StationSelectedMsg:
		station := gtfs.Stop(msg)
		m.selectedStation = station
		// Update departure table with new station's departures
		if !m.realtimeLoading {
			stopIds := []string{station.StopId+"N", station.StopId+"S"}
			m.departures = gtfs.FindDepartures(stopIds, m.realtime)
			m.departureTable = departuretable.NewModel(m.departures)
			m.departureTable.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
		}
	}

	// Update both components and batch their commands
	var cmds []tea.Cmd

	if !m.scheduleLoading {
		updatedModel, cmd := m.stationList.Update(msg)
		m.stationList = *updatedModel.(*stationlist.Model)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	if !m.realtimeLoading {
		updatedModel, cmd := m.departureTable.Update(msg)
		m.departureTable = *updatedModel.(*departuretable.Model)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *model) View() string {
	if m.scheduleLoading || m.realtimeLoading {
		return lipgloss.NewStyle().
			Width(m.width).
			Height(m.height).
			Align(lipgloss.Center, lipgloss.Center).
			Render(splash.Model{}.View())
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, m.stationList.View(), m.departureTable.View())
}

type gotScheduleMsg gtfs.Schedule

func getSchedule() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(500 * time.Millisecond)
		schedule := gtfs.GetSchedule()
		return gotScheduleMsg(schedule)
	}
}

type gotRealtimeMsg []gtfs.RealtimeFeed

func getRealtime() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(500 * time.Millisecond)
		feeds := gtfs.FetchFeeds()
		return gotRealtimeMsg(feeds)
	}
}
