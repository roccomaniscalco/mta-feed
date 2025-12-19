package gtfs

import (
	"log"
	"os"
	"strconv"
	"strings"
)

func CreateStopIdToName() map[string]string {
	stopIdToName := make(map[string]string)

	bytes, err := os.ReadFile(dataDir + stopsPath)
	if err != nil {
		log.Fatal(err)
	}

	content := string(bytes)
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		fields := strings.Split(line, ",")
		if len(fields) >= 2 {
			stopId := fields[0]
			stopName := fields[1]
			stopIdToName[stopId] = stopName
		}
	}

	return stopIdToName
}

type Stop struct {
	StopId        string
	StopName      string
	StopLat       float64
	StopLon       float64
	LocationType  string
	ParentStation string
}

func GetParentStations() []Stop {
	var parentStations []Stop

	bytes, err := os.ReadFile(dataDir + stopsPath)
	if err != nil {
		log.Fatal(err)
	}

	content := string(bytes)
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		fields := strings.Split(line, ",")

		if len(fields) < 6 {
			continue
		}
		
		StopId := fields[0]
		StopName := fields[1]
		StopLat, _ := strconv.ParseFloat(fields[2], 64)
		StopLon, _ := strconv.ParseFloat(fields[3], 64)
		LocationType := fields[4]
		ParentStation := fields[5]

		if LocationType == "1" {
			parentStation := Stop{
				StopId:        StopId,
				StopName:      StopName,
				StopLat:       StopLat,
				StopLon:       StopLon,
				LocationType:  LocationType,
				ParentStation: ParentStation,
			}
			parentStations = append(parentStations, parentStation)
		}
	}

	return parentStations
}
