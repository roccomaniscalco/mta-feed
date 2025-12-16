package main

import (
	"fmt"
	"io"
	"log"
	"maps"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"mta-feed/realtime"
)

const LOWEST_FILE_PERMS = 0644

const STOPS_FILE_PATH = "stops.txt"

const (
	ACE_FEED_URL       = "https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-ace"
	BDFM_FEED_URL      = "https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-bdfm"
	G_FEED_URL         = "https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-g"
	JZ_FEED_URL        = "https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-jz"
	NQRW_FEED_URL      = "https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-nqrw"
	L_FEED_URL         = "https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-l"
	_1234567S_FEED_URL = "https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs"
	SI_FEED_URL        = "https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-si"
)

type departure struct {
	stopId        string
	stopName      string
	routeId       string
	finalStopId   string
	finalStopName string
	time          int64
	delay         int32
}

var feedUrls = []string{
	ACE_FEED_URL,
	_1234567S_FEED_URL,
}

var stopIds = []string{
	"A46N",
	"A46S",
	"239N",
	"239S",
}

var stopIdToName = createStopIdToName()

func main() {
	feeds, err := fetchFeeds(feedUrls)
	if err != nil {
		log.Fatal("Error Fetching Feeds:", err)
	}

	departures := findDepartures(stopIds, feeds)
	printDepartures(departures)
}

func printDepartures(departures []departure) {
	currentTime := time.Now().Unix()

	// stopName -> routeId -> finalStopName -> times
	grouped := make(map[string]map[string]map[string][]int64)

	// Create nested map of departures
	for _, d := range departures {
		if grouped[d.stopName] == nil {
			grouped[d.stopName] = make(map[string]map[string][]int64)
		}
		if grouped[d.stopName][d.routeId] == nil {
			grouped[d.stopName][d.routeId] = make(map[string][]int64)
		}
		grouped[d.stopName][d.routeId][d.finalStopName] =
			append(grouped[d.stopName][d.routeId][d.finalStopName], d.time)
	}

	const (
		Bold  = "\033[1m"
		Reset = "\033[0m"
	)

	// Sort fields at each layer
	stopNames := slices.Sorted(maps.Keys(grouped))
	for _, stopName := range stopNames {
		routeIds := slices.Sorted(maps.Keys(grouped[stopName]))
		fmt.Printf("\n%s%s%s\n", Bold, stopName, Reset)

		for _, routeId := range routeIds {
			finalStopNames := slices.Sorted(maps.Keys(grouped[stopName][routeId]))
			fmt.Printf("\n%s Train\n", routeId)

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

				fmt.Printf("%s %v\n", finalStopName, minutesTilDepartures)
			}
		}
	}
}

func findDepartures(stopIds []string, feeds []*realtime.FeedMessage) []departure {
	departures := []departure{}

	for _, stopId := range stopIds {
		for _, feed := range feeds {
			for _, feedEntity := range feed.GetEntity() {
				tripUpdate := feedEntity.GetTripUpdate()
				stopTimes := tripUpdate.GetStopTimeUpdate()
				for _, stopTime := range stopTimes {
					if stopTime.GetStopId() == stopId {
						finalStopId := stopTimes[len(stopTimes)-1].GetStopId()
						departures = append(departures, departure{
							stopId:        stopId,
							stopName:      stopIdToName[stopId],
							routeId:       tripUpdate.Trip.GetRouteId(),
							finalStopId:   finalStopId,
							finalStopName: stopIdToName[finalStopId],
							time:          stopTime.GetDeparture().GetTime(),
							delay:         stopTime.GetDeparture().GetDelay(),
						})
					}
				}
			}
		}
	}

	return departures
}

func createStopIdToName() map[string]string {
	stopIdToName := make(map[string]string)

	bytes, err := os.ReadFile(STOPS_FILE_PATH)
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

func fetchFeeds(feedUrls []string) ([]*realtime.FeedMessage, error) {
	feeds := make([]*realtime.FeedMessage, len(feedUrls))
	var g errgroup.Group

	for i, feedUrl := range feedUrls {
		g.Go(func() error {
			i, feedUrl := i, feedUrl // capture loop variables

			feed, err := fetchFeed(feedUrl)
			if err != nil {
				return err
			}

			// writeFeed(feed)

			feeds[i] = feed
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}
	return feeds, nil
}

func fetchFeed(feedUrl string) (*realtime.FeedMessage, error) {
	resp, err := http.Get(feedUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	feed := &realtime.FeedMessage{}
	if err := proto.Unmarshal(body, feed); err != nil {
		return nil, err
	}

	return feed, nil
}

func writeFeed(msg *realtime.FeedMessage) {
	marshallOptions := protojson.MarshalOptions{
		Indent: "  ",
	}

	feedJson, err := marshallOptions.Marshal(msg)
	if err != nil {
		log.Fatal(err)
	}

	outFile := fmt.Sprintf("out/mta-feed-%d.json", *msg.Header.Timestamp)

	err = os.WriteFile(outFile, feedJson, LOWEST_FILE_PERMS)
	if err != nil {
		log.Fatal(err)
	}
}
