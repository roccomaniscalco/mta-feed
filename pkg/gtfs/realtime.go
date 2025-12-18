package gtfs

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"nyct-feed/proto/gtfs"
)

var feedUrls = [8]string{
	"https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-ace",
	"https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-bdfm",
	"https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-g",
	"https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-jz",
	"https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-nqrw",
	"https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-l",
	"https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs",
	"https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-si",
}

// Fetch feeds for all lines concurrently
func FetchFeeds() ([]*gtfs.FeedMessage, error) {
	feeds := make([]*gtfs.FeedMessage, len(feedUrls))
	var g errgroup.Group

	for i, feedUrl := range feedUrls {
		g.Go(func() error {
			i, feedUrl := i, feedUrl // capture loop variables

			feed, err := fetchFeed(feedUrl)
			if err != nil {
				return err
			}

			feeds[i] = feed
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}
	return feeds, nil
}

func fetchFeed(feedUrl string) (*gtfs.FeedMessage, error) {
	resp, err := http.Get(feedUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	feed := &gtfs.FeedMessage{}
	if err := proto.Unmarshal(body, feed); err != nil {
		return nil, err
	}

	return feed, nil
}

const (
	directoryPerms = 0755
	filePerms      = 0644
)

// Write feed messages to /out. Helpful for debugging
func writeFeed(msg *gtfs.FeedMessage) {
	marshallOptions := protojson.MarshalOptions{
		Indent: "  ",
	}

	feedJson, err := marshallOptions.Marshal(msg)
	if err != nil {
		log.Fatal(err)
	}

	err = os.MkdirAll("out", directoryPerms)
	if err != nil {
		log.Fatal(err)
	}

	outFile := fmt.Sprintf("out/mta-feed-%d.json", *msg.Header.Timestamp)

	err = os.WriteFile(outFile, feedJson, filePerms)
	if err != nil {
		log.Fatal(err)
	}
}
