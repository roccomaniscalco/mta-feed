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
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

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

func (m model) RouteBadge(routeShortName string, fg string, bg string) string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#"+fg)).
		Background(lipgloss.Color("#"+bg)).
		Bold(true).
		Padding(0, 1).
		MarginRight(1)

	return style.Render(routeShortName)
}

func main() {
	schedule, _ := gtfs.GetSchedule()
	stations := schedule.GetStations()

	m := model{}

	items := []list.Item{}
	for _, station := range stations {
		routeBadges := []string{}
		for _, route := range schedule.Routes {
			if _, exists := station.RouteIds[route.RouteId]; exists {
				badge := m.RouteBadge(route.RouteShortName, route.RouteTextColor, route.RouteColor)
				routeBadges = append(routeBadges, badge)
			}
		}

		routeIdsStr := lipgloss.JoinHorizontal(lipgloss.Left, routeBadges...)
		items = append(items, item{title: station.StopName, desc: routeIdsStr})
	}

	m.list = list.New(items, list.NewDefaultDelegate(), 0, 0)
	m.list.Title = "Stations"

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
