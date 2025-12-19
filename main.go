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
	err := gtfs.FetchAndStoreSchedule()
	if err != nil {
		log.Fatal("Error Fetching Schedule: ", err)
	}

	feeds, err := gtfs.FetchFeeds()
	if err != nil {
		log.Fatal("Error Fetching Feeds: ", err)
	}

	stopIdToName := gtfs.CreateStopIdToName()
	departures := gtfs.FindDepartures(stopIds, feeds)

	for i := range departures {
		departures[i].StopName = stopIdToName[departures[i].StopId]
		departures[i].FinalStopName = stopIdToName[departures[i].FinalStopId]
	}

	gtfs.PrintDepartures(departures)
}

