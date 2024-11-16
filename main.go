package main

import (
	"fmt"
	"io"
	"log"
	"mta-feed/gtfs-realtime"
	"net/http"

	"google.golang.org/protobuf/proto"
)

var STOP_ID = "L17S" // canarsie
// var STOP_ID = "L17N" // 8 av

func main() {
	msg, err := requestFeedMessage()
	if err != nil {
		log.Fatal(err)
	}

	msgTime := int64(*msg.Header.Timestamp)
	arrivalTimes := getArrivalTimesByStopId(msg, STOP_ID)

	for _, t := range arrivalTimes {
		fmt.Println((t - msgTime) / 60)
	}
}

func getArrivalTimesByStopId(msg *pb.FeedMessage, stopId string) []int64 {
	arrivalTimes := []int64{}

	for _, entity := range msg.GetEntity() {
		for _, update := range entity.GetTripUpdate().GetStopTimeUpdate() {
			if update.GetStopId() == stopId {
				arrivalTimes = append(arrivalTimes, update.GetArrival().GetTime())
			}
		}
	}

	return arrivalTimes
}

func requestFeedMessage() (*pb.FeedMessage, error) {
	endpointUri := "https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-l"

	resp, err := http.Get(endpointUri)
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
