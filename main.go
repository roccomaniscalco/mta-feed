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
	stopId string
}

var lines = []line{
	{
		feedUrl: L_URL,
		stops: []stop{
			{
				stopId: "L17N",
			},
			{
				stopId: "L17S",
			},
		},
	},
	{
		feedUrl: M_URL,
		stops: []stop{
			{
				stopId: "M08N",
			},
			{
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

		msgTime := int64(*msg.Header.Timestamp)

		for _, stop := range line.stops {
			fmt.Println()

			arrivalTimes := stop.getArrivalTimes(msg)[:2]
			nextArrivalTimes := arrivalTimes[:2]

			for _, arrivalTime := range nextArrivalTimes {
				minTilArrival := (arrivalTime - msgTime) / 60
				fmt.Println(minTilArrival)
			}
		}
	}
}

func (stop stop) getArrivalTimes(msg *pb.FeedMessage) []int64 {
	arrivalTimes := []int64{}

	for _, entity := range msg.GetEntity() {
		for _, update := range entity.GetTripUpdate().GetStopTimeUpdate() {
			if update.GetStopId() == stop.stopId {
				arrivalTimes = append(arrivalTimes, update.GetArrival().GetTime())
			}
		}
	}

	return arrivalTimes
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
	err = os.WriteFile("mta-feed.json", msgJson, LOWEST_FILE_PERMS)
	if err != nil {
		log.Fatal(err)
	}
}
