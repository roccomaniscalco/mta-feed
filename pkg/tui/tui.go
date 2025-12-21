package main

import (
	"fmt"
	"nyct-feed/pkg/gtfs"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type item struct {
	station     gtfs.Stop
	routeBadges string
}

func (i item) Title() string       { return i.station.StopName }
func (i item) Description() string { return i.routeBadges }
func (i item) FilterValue() string { return i.station.StopName }

type model struct {
	list list.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return docStyle.Render(m.list.View())
}

func RouteBadge(route gtfs.Route) string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#"+route.RouteTextColor)).
		Background(lipgloss.Color("#"+route.RouteColor)).
		Bold(true).
		Padding(0, 1).
		MarginRight(1)

	return style.Render(route.RouteShortName)
}

func main() {
	schedule := gtfs.GetSchedule()
	stations := schedule.GetStations()

	items := []list.Item{}
	for _, station := range stations {
		routeBadges := []string{}
		for _, route := range schedule.Routes {
			if _, exists := station.RouteIds[route.RouteId]; exists {
				routeBadges = append(routeBadges, RouteBadge(route))
			}
		}
		routeBadgesStr := lipgloss.JoinHorizontal(lipgloss.Left, routeBadges...)
		items = append(items, item{station: station, routeBadges: routeBadgesStr})
	}

	list := list.New(items, list.NewDefaultDelegate(), 0, 0)
	list.Title = "Stations"
	list.SetShowPagination(false)

	m := model{
		list: list,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
