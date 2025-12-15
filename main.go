package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"mta-feed/realtime"
)

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
	routeId     string
	finalStopId string
	time        int64
}

var feedUrls = []string{
	L_FEED_URL,
	BDFM_FEED_URL,
}

var stopIds = []string{
	"L17N",
	"L17S",
	"M08N",
	"M08S",
}

func main() {
	arrivalsByStopId := make(map[string][]arrival)

	for _, feedUrl := range feedUrls {
		feedMessage, err := requestFeedMessage(feedUrl)
		if err != nil {
			log.Fatal(err)
		}

		writeFeedMessage(feedMessage)

		for _, stopId := range stopIds {
			arrivals := findArrivals(stopId, feedMessage)
			arrivalsByStopId[stopId] = append(arrivalsByStopId[stopId], arrivals...)
		}
	}

	fmt.Println(arrivalsByStopId)
}

func findArrivals(stopId string, feedMessage *realtime.FeedMessage) []arrival {
	arrivals := []arrival{}

	for _, feedEntity := range feedMessage.GetEntity() {
		tripUpdate := feedEntity.GetTripUpdate()
		stopTimeUpdates := tripUpdate.GetStopTimeUpdate()
		for _, stopTimeUpdate := range stopTimeUpdates {
			if stopTimeUpdate.GetStopId() == stopId {
				arrivals = append(arrivals, arrival{
					routeId:     tripUpdate.Trip.GetRouteId(),
					finalStopId: stopTimeUpdates[len(stopTimeUpdates)-1].GetStopId(),
					time:        stopTimeUpdate.GetArrival().GetTime(),
				})
			}
		}
	}

	return arrivals
}

func requestFeedMessage(feedUrl string) (*realtime.FeedMessage, error) {
	resp, err := http.Get(feedUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	msg := &realtime.FeedMessage{}
	if err := proto.Unmarshal(body, msg); err != nil {
		return nil, err
	}

	return msg, nil
}

func writeFeedMessage(msg *realtime.FeedMessage) {
	marshallOptions := protojson.MarshalOptions{
		Indent: "  ",
	}

	msgJson, err := marshallOptions.Marshal(msg)
	if err != nil {
		log.Fatal(err)
	}

	const (
		LOWEST_FILE_PERMS = 0644
	)

	outFile := fmt.Sprintf("out/mta-feed-%d.json", *msg.Header.Timestamp)

	err = os.WriteFile(outFile, msgJson, LOWEST_FILE_PERMS)
	if err != nil {
		log.Fatal(err)
	}
}
