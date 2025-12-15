package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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

type arrival struct {
	stopId      string
	routeId     string
	finalStopId string
	time        int64
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

func main() {
	feeds, err := fetchFeeds(feedUrls)
	if err != nil {
		log.Fatal("Error Fetching Feeds:", err)
	}
	
	stopIdToName := createStopIdToName()
	arrivals := findArrivals(stopIds, feeds)
	currentTime := time.Now().Unix()

	for _, arrival := range arrivals {
		fmt.Printf("%s %s %s %v\n",
			arrival.stopId,
			arrival.routeId,
			stopIdToName[arrival.finalStopId],
			(arrival.time-currentTime)/60,
		)
	}
}

func findArrivals(stopIds []string, feeds []*realtime.FeedMessage) []arrival {
	arrivals := []arrival{}

	for _, stopId := range stopIds {
		for _, feed := range feeds {
			for _, feedEntity := range feed.GetEntity() {
				tripUpdate := feedEntity.GetTripUpdate()
				stopTimeUpdates := tripUpdate.GetStopTimeUpdate()
				for _, stopTimeUpdate := range stopTimeUpdates {
					if stopTimeUpdate.GetStopId() == stopId {
						arrivals = append(arrivals, arrival{
							stopId:      stopId,
							routeId:     tripUpdate.Trip.GetRouteId(),
							finalStopId: stopTimeUpdates[len(stopTimeUpdates)-1].GetStopId(),
							time:        stopTimeUpdate.GetArrival().GetTime(),
						})
					}
				}
			}
		}
	}

	return arrivals
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

			writeFeed(feed)

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
