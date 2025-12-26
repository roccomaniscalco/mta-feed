package main

import (
	// "log"
	"log"
	"nyct-feed/pkg/gtfs"
	// "nyct-feed/pkg/tui"
	// tea "github.com/charmbracelet/bubbletea"
)

var stopIds = []string{
	"A46N",
	"A46S",
	"239N",
	"239S",
}

func main() {
	feeds := gtfs.FetchFeeds()
	schedule := gtfs.GetSchedule()
	stopIdToName := schedule.GetStopIdToName()

	departures := gtfs.FindDepartures(stopIds, feeds, stopIdToName)
	log.Println(departures)

	// m := tui.NewModel()
	// p := tea.NewProgram(&m, tea.WithAltScreen())
	// if _, err := p.Run(); err != nil {
	// 	log.Fatalf("Error running program:", err)
	// }
}
