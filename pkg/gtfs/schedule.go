package gtfs

import (
	"archive/zip"
	"bytes"
	"io"
	"net/http"
	"os"
)

const scheduleUrl = "https://rrgtfsfeeds.s3.amazonaws.com/gtfs_subway.zip"

var targetFiles = []string{
	routesPath,
	stopsPath,
	shapesPath,
}

// Fetch and persist GTFS schedule files
func FetchAndStoreSchedule() error {
	files, err := fetchSchedule(targetFiles)
	if err != nil {
		return err
	}

	for _, file := range files {
		if err := storeSchedule(file); err != nil {
			return err
		}
	}

	return nil
}

func fetchSchedule(fileNames []string) ([]*zip.File, error) {
	// Download the ZIP file
	resp, err := http.Get(scheduleUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the ZIP data into memory
	zipData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Create a ZIP reader
	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return nil, err
	}

	// Gather the target files
	files := []*zip.File{}
	for _, file := range zipReader.File {
		for _, targetFile := range targetFiles {
			if file.Name == targetFile {
				files = append(files, file)
			}
		}
	}

	return files, nil
}

func storeSchedule(file *zip.File) error {
	rc, err := file.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return err
	}

	// Ensure the data directory exists
	if err := os.MkdirAll(dataDir, dirPerms); err != nil {
		return err
	}

	err = os.WriteFile(dataDir+file.Name, data, filePerms)
	if err != nil {
		return err
	}

	return nil
}
