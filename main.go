package main

import (
	"fmt"
	"io"
	"log"
	"mta-feed/gtfs-realtime"
	"net/http"

	"google.golang.org/protobuf/proto"
)

func main() {
	msg, err := getGtfsRealtime()
	if err != nil {
		log.Fatal(err)
	}

	for _, entity := range msg.GetEntity() {
		fmt.Printf("Entity: %+v\n", entity.GetTripUpdate())
	}
}

func getGtfsRealtime() (*pb.FeedMessage, error) {
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
