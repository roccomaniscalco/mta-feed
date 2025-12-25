package gtfs

import (
	"fmt"
	"maps"
	"slices"
	"time"
)

type Departure struct {
	StopId        string
	StopName      string
	RouteId       string
	FinalStopId   string
	FinalStopName string
	Time          int64
	Delay         int32
}

func FindDepartures(stopIds []string, feeds []RealtimeFeed) []Departure {
	departures := []Departure{}

	for _, stopId := range stopIds {
		for _, feed := range feeds {
			for _, feedEntity := range feed.GetEntity() {
				tripUpdate := feedEntity.GetTripUpdate()
				stopTimes := tripUpdate.GetStopTimeUpdate()
				for _, stopTime := range stopTimes {
					if stopTime.GetStopId() == stopId {
						finalStopId := stopTimes[len(stopTimes)-1].GetStopId()
						departures = append(departures, Departure{
							StopId:        stopId,
							RouteId:       tripUpdate.Trip.GetRouteId(),
							FinalStopId:   finalStopId,
							Time:          stopTime.GetDeparture().GetTime(),
							Delay:         stopTime.GetDeparture().GetDelay(),
						})
					}
				}
			}
		}
	}

	return departures
}

var routeIdToColor = map[string]string{
	"A":  "\033[38;2;0;98;207m",    // #0062CF - Blue
	"C":  "\033[38;2;0;98;207m",    // #0062CF - Blue
	"E":  "\033[38;2;0;98;207m",    // #0062CF - Blue
	"B":  "\033[38;2;235;104;0m",   // #EB6800 - Orange
	"D":  "\033[38;2;235;104;0m",   // #EB6800 - Orange
	"F":  "\033[38;2;235;104;0m",   // #EB6800 - Orange
	"FX": "\033[38;2;235;104;0m",   // #EB6800 - Orange
	"M":  "\033[38;2;235;104;0m",   // #EB6800 - Orange
	"G":  "\033[38;2;121;149;52m",  // #799534 - Green
	"J":  "\033[38;2;142;92;51m",   // #8E5C33 - Brown
	"Z":  "\033[38;2;142;92;51m",   // #8E5C33 - Brown
	"L":  "\033[38;2;124;133;140m", // #7C858C - Gray
	"N":  "\033[38;2;246;188;38m",  // #F6BC26 - Yellow
	"Q":  "\033[38;2;246;188;38m",  // #F6BC26 - Yellow
	"R":  "\033[38;2;246;188;38m",  // #F6BC26 - Yellow
	"W":  "\033[38;2;246;188;38m",  // #F6BC26 - Yellow
	"GS": "\033[38;2;124;133;140m", // #7C858C - Gray
	"FS": "\033[38;2;124;133;140m", // #7C858C - Gray
	"H":  "\033[38;2;124;133;140m", // #7C858C - Gray
	"1":  "\033[38;2;216;34;51m",   // #D82233 - Red
	"2":  "\033[38;2;216;34;51m",   // #D82233 - Red
	"3":  "\033[38;2;216;34;51m",   // #D82233 - Red
	"4":  "\033[38;2;0;153;82m",    // #009952 - Green
	"5":  "\033[38;2;0;153;82m",    // #009952 - Green
	"6":  "\033[38;2;0;153;82m",    // #009952 - Green
	"6X": "\033[38;2;0;153;82m",    // #009952 - Green
	"7":  "\033[38;2;154;56;161m",  // #9A38A1 - Purple
	"7X": "\033[38;2;154;56;161m",  // #9A38A1 - Purple
	"SI": "\033[38;2;8;23;156m",    // #08179C - Dark Blue
}

const (
	bold  = "\033[1m"
	reset = "\033[0m"
)

func PrintDepartures(departures []Departure) {
	currentTime := time.Now().Unix()

	// stopName -> routeId -> finalStopName -> times
	grouped := make(map[string]map[string]map[string][]int64)

	// Create nested map of departures
	for _, d := range departures {
		if grouped[d.StopName] == nil {
			grouped[d.StopName] = make(map[string]map[string][]int64)
		}
		if grouped[d.StopName][d.RouteId] == nil {
			grouped[d.StopName][d.RouteId] = make(map[string][]int64)
		}
		grouped[d.StopName][d.RouteId][d.FinalStopName] =
			append(grouped[d.StopName][d.RouteId][d.FinalStopName], d.Time)
	}

	// Sort fields at each layer
	stopNames := slices.Sorted(maps.Keys(grouped))
	for _, stopName := range stopNames {
		routeIds := slices.Sorted(maps.Keys(grouped[stopName]))
		fmt.Printf("\n%s%s%s\n", bold, stopName, reset)

		for _, routeId := range routeIds {
			finalStopNames := slices.Sorted(maps.Keys(grouped[stopName][routeId]))
			fmt.Println()

			for _, finalStopName := range finalStopNames {
				times := (grouped[stopName][routeId][finalStopName])
				slices.Sort(times)

				// Display 2 soonest departure times as countdown in minutes
				minutesTilDepartures := []int64{}
				for _, time := range times {
					minutes := (time - currentTime) / 60
					if minutes >= 0 && len(minutesTilDepartures) < 2 {
						minutesTilDepartures = append(minutesTilDepartures, minutes)
					}
				}

				fmt.Printf("%s%s%s %s %v\n", routeIdToColor[routeId], routeId, reset, finalStopName, minutesTilDepartures)
			}
		}
	}
}
