package main

import (
	"log"

	"nyct-feed/pkg/gtfs"
)

var stopIds = []string{
	"A46N",
	"A46S",
	"239N",
	"239S",
}

func main() {
	feeds, err := gtfs.FetchFeeds()
	if err != nil {
		log.Fatal("Error Fetching Feeds:", err)
	}

	departures := gtfs.FindDepartures(stopIds, feeds)
	gtfs.PrintDepartures(departures)
}

