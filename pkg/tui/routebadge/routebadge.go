package routebadge

import (
	"nyct-feed/pkg/gtfs"

	"github.com/charmbracelet/lipgloss"
)

func Render(route gtfs.Route) string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#"+route.RouteTextColor)).
		Background(lipgloss.Color("#"+route.RouteColor)).
		Bold(true).
		Padding(0, 1)

	return style.Render(route.RouteShortName)
}
