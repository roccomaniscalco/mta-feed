package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"mta-feed/gtfs-realtime"
	"net/http"
	"google.golang.org/protobuf/proto"
)

func main() {
	endpointUri := "https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-l"
	resp, err := http.Get(endpointUri)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	msg := &pb.FeedMessage{}
	if err := proto.Unmarshal(body, msg); err != nil {
		log.Fatalf("Failed to unmarshal response: %v", err)
	}

	for _, entity := range msg.GetEntity() {
		fmt.Printf("Entity: %+v\n", entity.GetTripUpdate())
	}
}

