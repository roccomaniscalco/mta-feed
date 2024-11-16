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
	L_URL = "https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-l"
	M_URL = "https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-bdfm"
)

type line struct {
	feedUrl string
	stops   []stop
}

type stop struct {
	stopId            string
	headsign          string
	delay             bool
	arrivalCountdowns []int64
}

var lines = []line{
	{
		feedUrl: L_URL,
		stops: []stop{
			{
				headsign: "8th Ave",
				stopId: "L17N",
			},
			{
				headsign: "Canarsie",
				stopId: "L17S",
			},
		},
	},
	{
		feedUrl: M_URL,
		stops: []stop{
			{
				headsign: "Forest Hills",
				stopId: "M08N",
			},
			{
				headsign: "Middle Village",
				stopId: "M08S",
			},
		},
	},
}

func main() {
	for _, line := range lines {
		msg, err := line.requestFeedMessage()
		if err != nil {
			log.Fatal(err)
		}

		writeFeedMessage(msg)

		for _, stop := range line.stops {
			stop.processMsg(msg)
			stop.print()
		}
	}
}

func (stop *stop) processMsg(msg *pb.FeedMessage) {
	msgTime := int64(*msg.Header.Timestamp)

	arrivals := []*pb.TripUpdate_StopTimeEvent{}
	for _, entity := range msg.GetEntity() {
		for _, update := range entity.GetTripUpdate().GetStopTimeUpdate() {
			if update.GetStopId() == stop.stopId {
				arrivals = append(arrivals, update.GetArrival())
			}
		}
	}

	for _, arrival := range arrivals {
		arrivalCountdown := (arrival.GetTime() - msgTime) / 60
		stop.arrivalCountdowns = append(stop.arrivalCountdowns, arrivalCountdown)
		if arrival.GetDelay() > 0 {
			stop.delay = true
		}
	}
}

func (stop stop) print() {
	var delayStr string
	if stop.delay {
		delayStr = "(delayed)"
	}

	fmt.Printf("%s: %v %s\n", stop.headsign, stop.arrivalCountdowns[:2], delayStr)
}

func (line line) requestFeedMessage() (*pb.FeedMessage, error) {
	resp, err := http.Get(line.feedUrl)
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

	outFile := fmt.Sprintf("mta-feed-%d.json", *msg.Header.Timestamp)

	err = os.WriteFile(outFile, msgJson, LOWEST_FILE_PERMS)
	if err != nil {
		log.Fatal(err)
	}
}
