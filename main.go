package main

import (
	"fmt"
	"io"
	"log"
	"mta-feed/pb"
	"net/http"
	"os"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const (
	L_URL    = "https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-l"
	BDFM_URL = "https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-bdfm"
)

type subway struct {
	feedUrl   string
	platforms []platform
}

type platform struct {
	stopId            string
	routeId           string
	headsign          string
	delay             bool
	arrivalCountdowns []int64
}

var subways = []subway{
	{
		feedUrl: L_URL,
		platforms: []platform{
			{
				headsign: "8th Av",
				stopId:   "L17N",
			},
			{
				headsign: "Canarsie",
				stopId:   "L17S",
			},
		},
	},
	{
		feedUrl: BDFM_URL,
		platforms: []platform{
			{
				headsign: "Myrtle Av",
				stopId:   "M08N",
			},
			{
				headsign: "Middle Village",
				stopId:   "M08S",
			},
		},
	},
}

func main() {
	for _, subway := range subways {
		msg, err := subway.requestFeedMessage()
		if err != nil {
			log.Fatal(err)
		}

		writeFeedMessage(msg)

		for _, platform := range subway.platforms {
			platform.processMsg(msg)
			platform.print()
		}
	}
}

func (platform *platform) processMsg(msg *pb.FeedMessage) {
	msgTime := int64(*msg.Header.Timestamp)

	arrivals := []*pb.TripUpdate_StopTimeEvent{}

	for _, entity := range msg.GetEntity() {
		tripUpdate := entity.GetTripUpdate()
		for _, stopTimeUpdate := range tripUpdate.GetStopTimeUpdate() {
			if stopTimeUpdate.GetStopId() == platform.stopId {
				platform.routeId = tripUpdate.GetTrip().GetRouteId()
				arrivals = append(arrivals, stopTimeUpdate.GetArrival())
			}
		}
	}

	for _, arrival := range arrivals {
		arrivalCountdown := (arrival.GetTime() - msgTime) / 60
		if len(platform.arrivalCountdowns) < 2 {
			platform.arrivalCountdowns = append(platform.arrivalCountdowns, arrivalCountdown)
		}
		if arrival.GetDelay() > 0 {
			platform.delay = true
		}
	}
}

func (platform platform) print() {
	var delayStr string
	if platform.delay {
		delayStr = "(delayed)"
	}

	fmt.Printf("%s %s %v %s\n", platform.routeId, platform.headsign, platform.arrivalCountdowns, delayStr)
}

func (subway subway) requestFeedMessage() (*pb.FeedMessage, error) {
	resp, err := http.Get(subway.feedUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	msg := &pb.FeedMessage{}
	if err := proto.Unmarshal(body, msg); err != nil {
		return nil, err
	}

	return msg, nil
}

func writeFeedMessage(msg *pb.FeedMessage) {
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
