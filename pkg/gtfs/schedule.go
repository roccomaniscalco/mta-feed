package gtfs

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const scheduleUrl = "https://rrgtfsfeeds.s3.amazonaws.com/gtfs_subway.zip"

type Schedule struct {
	Stops     []Stop     `file:"stops.txt"`
	StopTimes []StopTime `file:"stop_times.txt"`
	Trips     []Trip     `file:"trips.txt"`
	Routes    []Route    `file:"routes.txt"`
	Shapes    []Shape    `file:"shapes.txt"`
}

type Stop struct {
	StopId        string  `csv:"stop_id"`
	StopName      string  `csv:"stop_name"`
	StopLat       float64 `csv:"stop_lat"`
	StopLon       float64 `csv:"stop_lon"`
	LocationType  int     `csv:"location_type"`
	ParentStation string  `csv:"parent_station"`
	RouteIds      map[string]bool
}

type StopTime struct {
	TripId        string `csv:"trip_id"`
	StopId        string `csv:"stop_id"`
	ArrivalTime   string `csv:"arrival_time"`
	DepartureTime string `csv:"departure_time"`
	StopSequence  int    `csv:"stop_sequence"`
}

type Trip struct {
	RouteId      string `csv:"route_id"`
	TripId       string `csv:"trip_id"`
	ServiceId    string `csv:"service_id"`
	TripHeadsign string `csv:"trip_headsign"`
	DirectionId  int    `csv:"direction_id"`
	ShapeId      string `csv:"shape_id"`
}

type Route struct {
	RouteId        string `csv:"route_id"`
	AgencyId       string `csv:"agency_id"`
	RouteShortName string `csv:"route_short_name"`
	RouteLongName  string `csv:"route_long_name"`
	RouteDesc      string `csv:"route_desc"`
	RouteType      int    `csv:"route_type"`
	RouteUrl       string `csv:"route_url"`
	RouteColor     string `csv:"route_color"`
	RouteTextColor string `csv:"route_text_color"`
	RouteSortOrder int    `csv:"route_sort_order"`
}

type Shape struct {
	ShapeId    string  `csv:"shape_id"`
	ShapePtSeq int     `csv:"shape_pt_sequence"`
	ShapePtLat float64 `csv:"shape_pt_lat"`
	ShapePtLon float64 `csv:"shape_pt_lon"`
}

func (s *Schedule) GetStations() []Stop {
	var parentStations []Stop
	for _, stop := range s.Stops {
		if stop.LocationType == 1 {
			parentStations = append(parentStations, stop)
		}
	}

	parentStopIdToTripIds := make(map[string][]string)
	for _, stopTime := range s.StopTimes {
		parentStopId := stopTime.StopId[:3]
		parentStopIdToTripIds[parentStopId] =
			append(parentStopIdToTripIds[parentStopId], stopTime.TripId)
	}

	for i, station := range parentStations {
		stopTripIds := parentStopIdToTripIds[station.StopId]
		for _, stopTripId := range stopTripIds {
			for _, trip := range s.Trips {
				if stopTripId == trip.TripId {
					if parentStations[i].RouteIds == nil {
						parentStations[i].RouteIds = map[string]bool{}
					}
					parentStations[i].RouteIds[trip.RouteId] = true
				}
			}
		}
	}

	return parentStations
}

func CreateStopIdToName(stops []Stop) map[string]string {
	stopIdToName := make(map[string]string)

	for _, stop := range stops {
		stopIdToName[stop.StopId] = stop.StopName
	}

	return stopIdToName
}

// GetSchedule returns a GTFS Schedule struct containing all static files.
// The Schedule is requested then stored when missing or stale.
// Otherwise its files are read and parsed from storage.
func GetSchedule() (*Schedule, error) {
	var schedule Schedule
	scheduleType := reflect.TypeOf(schedule)

	currentTime := time.Now()
	isScheduleDirty := false

	// Schedule is dirty if any file is missing or older than 24 hrs
	for i := 0; i < scheduleType.NumField(); i++ {
		fileName := scheduleType.Field(i).Tag.Get("file")

		fileInfo, err := os.Stat(dataDir + fileName)
		if err != nil {
			isScheduleDirty = true
			break
		}
		if currentTime.Sub(fileInfo.ModTime()).Hours() > 24 {
			isScheduleDirty = true
			break
		}
	}

	if isScheduleDirty {
		if err := fetchAndStoreSchedule(); err != nil {
			return nil, err
		}
	}

	// Parse and set each item on schedule
	for i := 0; i < scheduleType.NumField(); i++ {
		field := scheduleType.Field(i)
		fileName := field.Tag.Get("file")
		fileRowType := field.Type.Elem()
		log.Println(fileName, fileRowType)

		bytes, err := os.ReadFile(dataDir + fileName)
		if err != nil {
			return nil, err
		}

		switch fileRowType {
		case reflect.TypeOf(Stop{}):
			rows, _ := parseCSV(bytes, Stop{})
			schedule.Stops = rows
		case reflect.TypeOf(StopTime{}):
			rows, _ := parseCSV(bytes, StopTime{})
			schedule.StopTimes = rows
		case reflect.TypeOf(Trip{}):
			rows, _ := parseCSV(bytes, Trip{})
			schedule.Trips = rows
		case reflect.TypeOf(Route{}):
			rows, _ := parseCSV(bytes, Route{})
			schedule.Routes = rows
		case reflect.TypeOf(Shape{}):
			rows, _ := parseCSV(bytes, Shape{})
			schedule.Shapes = rows
		}
	}

	return &schedule, nil
}

// parseCSV accept CSV bytes and parses each row into a struct of type R.
// Each field in R to be parsed must specify a csv tag denoting the column header.
// CSVError is returned if file could not be parsed.
func parseCSV[R any](bytes []byte, row R) ([]R, error) {
	r := reflect.TypeOf(row)
	rows := []R{}

	// Only accept structs
	if r.Kind() != reflect.Struct {
		return nil, errors.New("parseCSV must be passed a struct for R")
	}

	lines := strings.Split(string(bytes), "\n")
	headers := parseCSVLine(lines[0])

	// Map header to column number
	headerToCol := make(map[string]int)
	for i, header := range headers {
		headerToCol[header] = i
	}

	// Process data rows
	for _, line := range lines[1:] {
		// Skip blank lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		cells := parseCSVLine(line)
		newValue := reflect.New(r).Elem()

		// Populate fields based on CSV cells
		for i := 0; i < r.NumField(); i++ {
			field := newValue.Field(i)
			fieldType := r.Field(i)

			// Get header from tag or field name
			header := fieldType.Tag.Get("csv")

			// Skip fields missing the csv tag
			if header == "" {
				continue
			}

			// Find the corresponding col for header
			if col, exists := headerToCol[header]; exists && col < len(cells) {
				cell := cells[col]

				// Coerce cell to the correct field type and set it
				if field.CanSet() {
					switch fieldType.Type.Kind() {
					case reflect.String:
						val := strings.Trim(cell, "\"")
						field.SetString(val)
					case reflect.Int:
						val, _ := strconv.Atoi(cell)
						field.SetInt(int64(val))
					case reflect.Bool:
						val, _ := strconv.ParseBool(cell)
						field.SetBool(val)
					case reflect.Float64:
						val, _ := strconv.ParseFloat(cell, 64)
						field.SetFloat(val)
					}
				}
			}
		}

		rows = append(rows, newValue.Interface().(R))
	}

	return rows, nil
}

func parseCSVLine(line string) []string {
	cells := []string{}
	chars := strings.Split(line, "")
	cellStart := 0
	inQuotes := false

	for i, char := range chars {
		if char == "\"" {
			inQuotes = !inQuotes
		} else if !inQuotes && char == "," {
			field := strings.Join(chars[cellStart:i], "")
			cells = append(cells, field)
			cellStart = i + 1
		}
	}

	// Handle the last field (after final comma or whole line if no commas)
	if cellStart < len(chars) {
		field := strings.Join(chars[cellStart:], "")
		cells = append(cells, field)
	}

	return cells
}

// fetchAndStoreSchedule requests a schedule zip file and stores its contents
func fetchAndStoreSchedule() error {
	files, err := fetchSchedule()
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

func fetchSchedule() ([]*zip.File, error) {
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

	// Gather the files
	files := []*zip.File{}
	for _, file := range zipReader.File {
		files = append(files, file)
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
