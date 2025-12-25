package tui

import (
	"nyct-feed/pkg/gtfs"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type item struct {
	station     gtfs.Stop
	routeBadges string
}

func (i item) Title() string       { return i.station.StopName }
func (i item) Description() string { return i.routeBadges }
func (i item) FilterValue() string { return i.station.StopName }

type Model struct {
	list         list.Model
	selectedItem item
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetHeight(msg.Height - 2)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)

	// Keep a snapshot of the last selected item to preserve while filtering
	if !m.list.SettingFilter() {
		if item, ok := m.list.SelectedItem().(item); ok {
			m.selectedItem = item
		}
	}

	if m.list.SettingFilter() {
		m.list.Styles.TitleBar = m.list.Styles.TitleBar.
			BorderForeground(lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"})
	} else {
		m.list.Styles.TitleBar = m.list.Styles.TitleBar.
			BorderForeground(lipgloss.AdaptiveColor{Light: "#C2B8C2", Dark: "#4D4D4D"})
	}

	style := lipgloss.NewStyle().
		UnsetBackground().
		Foreground(lipgloss.AdaptiveColor{Light: "#A49FA5", Dark: "#777777"})
	if m.list.FilterValue() == "" {
		m.list.Title = Kbd("/") + style.Render("Search Stations")
	} else {
		m.list.Title = Kbd("/") + style.Render(m.list.FilterValue())
	}

	return m, cmd
}

func (m Model) View() string {
	return m.list.View()
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

func Kbd(key string) string {
	style := lipgloss.NewStyle().
		MarginRight(1).
		Foreground(lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#dddddd"})

	return style.Render(key)
}

func NewModel(stations []gtfs.Stop, routes []gtfs.Route, height int) Model {
	items := []list.Item{}
	for _, station := range stations {
		routeBadges := []string{}
		for _, route := range routes {
			if _, exists := station.RouteIds[route.RouteId]; exists {
				routeBadges = append(routeBadges, RouteBadge(route))
			}
		}
		routeBadgesStr := lipgloss.JoinHorizontal(lipgloss.Left, routeBadges...)
		items = append(items, item{station: station, routeBadges: routeBadgesStr})
	}

	list := list.New(items, list.NewDefaultDelegate(), 0, 0)
	list.SetWidth(40)
	list.SetHeight(height)

	list.SetShowPagination(false)
	list.SetShowHelp(false)
	list.SetShowStatusBar(false)

	list.Styles.TitleBar = lipgloss.NewStyle().
		Width(40).
		Padding(0, 1).
		MarginBottom(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.AdaptiveColor{Light: "#C2B8C2", Dark: "#4D4D4D"})

	list.Styles.Title = lipgloss.NewStyle().
		UnsetBackground().
		Foreground(lipgloss.AdaptiveColor{Light: "#A49FA5", Dark: "#777777"})

	list.FilterInput.Prompt = Kbd("/")
	list.FilterInput.Placeholder = "Search Stations"
	list.FilterInput.TextStyle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#dddddd"})

	return Model{
		list:         list,
		selectedItem: items[0].(item),
	}
}
