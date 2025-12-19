package gtfs

import (
	"log"
	"os"
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
