package static

import (
	"errors"
	"io"
	"net/http"
	"os"

	"github.com/patrickbr/gtfsparser"
)

const (
	MTA_STATIC_GTFS_URL      = "http://web.mta.info/developers/data/nyct/subway/google_transit.zip"
	MTA_STATIC_GTFS_OUT_PATH = "google_transit.zip"
)

func requestFeed() error {
	resp, err := http.Get(MTA_STATIC_GTFS_URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(MTA_STATIC_GTFS_OUT_PATH)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func GetFeed() (*gtfsparser.Feed, error) {
	feed := gtfsparser.NewFeed()

	if _, err := os.Stat(MTA_STATIC_GTFS_OUT_PATH); err == nil {
		err = feed.Parse(MTA_STATIC_GTFS_OUT_PATH)
		return feed, err

	} else if errors.Is(err, os.ErrNotExist) {
		err = requestFeed()
		if err == nil {
			err = feed.Parse(MTA_STATIC_GTFS_OUT_PATH)
		}
		return feed, err

	} else {
		return feed, err
	}
}
