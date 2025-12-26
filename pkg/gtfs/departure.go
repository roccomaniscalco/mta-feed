package gtfs

import (
	"slices"
)

type Departure struct {
	RouteId       string
	FinalStopId   string
	FinalStopName string
	Times         []int64
}

func FindDepartures(stopIds []string, feeds []RealtimeFeed, stopIdToName map[string]string) []Departure {
	tripToTimes := map[[2]string][]int64{}

	for _, stopId := range stopIds {
		for _, feed := range feeds {
			for _, feedEntity := range feed.GetEntity() {
				tripUpdate := feedEntity.GetTripUpdate()
				stopTimes := tripUpdate.GetStopTimeUpdate()
				routeId := tripUpdate.GetTrip().GetRouteId()
				for _, stopTime := range stopTimes {
					finalStopId := stopTimes[len(stopTimes)-1].GetStopId()
					tripKey := [2]string{routeId, finalStopId}
					// Exclude trips terminating at the target stop
					if stopTime.GetStopId() == stopId && finalStopId != stopId {
						tripToTimes[tripKey] = append(tripToTimes[tripKey], stopTime.GetDeparture().GetTime())
					}
				}
			}
		}
	}

	departures := []Departure{}
	for tripKey, times := range tripToTimes {
		routeId, finalStopId := tripKey[0], tripKey[1]
		slices.Sort(times)
		departures = append(departures, Departure{
			RouteId:       routeId,
			FinalStopId:   finalStopId,
			FinalStopName: stopIdToName[finalStopId],
			Times:         times,
		})
	}

	return departures
}
