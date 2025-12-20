package gtfs

import (
	"fmt"
	"log"
	"math"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type Stop struct {
	StopId        string  `csv:"stop_id"`
	StopName      string  `csv:"stop_name"`
	StopLat       float64 `csv:"stop_lat"`
	StopLon       float64 `csv:"stop_lon"`
	LocationType  int     `csv:"location_type"`
	ParentStation string  `csv:"parent_station"`
	RouteIds      map[string]bool
}

type Shape struct {
	ShapeId    string  `csv:"shape_id"`
	ShapePtSeq int     `csv:"shape_pt_sequence"`
	ShapePtLat float64 `csv:"shape_pt_lat"`
	ShapePtLon float64 `csv:"shape_pt_lon"`
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

func CreateStopIdToName() map[string]string {
	stopIdToName := make(map[string]string)

	stops := getStops()

	for _, stop := range stops {
		stopIdToName[stop.StopId] = stop.StopName
	}

	return stopIdToName
}

func createStopIdToRouteIds(stops []Stop, shapes []Shape) map[string]map[string]bool {
	stopIdToRouteIds := make(map[string]map[string]bool)

	for _, stop := range stops {
		for _, shape := range shapes {
			closeLat := math.Abs(stop.StopLat-shape.ShapePtLat) < 0.0001
			closeLon := math.Abs(stop.StopLon-shape.ShapePtLon) < 0.0001
			if closeLat && closeLon {
				routeId := strings.Split(shape.ShapeId, ".")[0]
				if stopIdToRouteIds[stop.StopId] == nil {
					stopIdToRouteIds[stop.StopId] = make(map[string]bool)
				}
				stopIdToRouteIds[stop.StopId][routeId] = true
			}
		}
	}

	return stopIdToRouteIds
}

func GetParentStations() []Stop {
	var parentStations []Stop

	stops := getStops()
	shapes := getShapes()

	for _, stop := range stops {
		if stop.LocationType == 1 {
			parentStations = append(parentStations, stop)
		}
	}

	stopIdToRouteIds := createStopIdToRouteIds(parentStations, shapes)

	for i, station := range parentStations {
		if routeIds, exists := stopIdToRouteIds[station.StopId]; exists {
			parentStations[i].RouteIds = routeIds
		}
	}

	return parentStations
}

func getStops() []Stop {
	stops, err := parseCSV[Stop](dataDir + stopsPath)
	if err != nil {
		log.Fatal(err)
	}
	return stops
}

func getShapes() []Shape {
	shapes, err := parseCSV[Shape](dataDir + shapesPath)
	if err != nil {
		log.Fatal(err)
	}
	return shapes
}

func GetRoutes() []Route {
	routes, err := parseCSV[Route](dataDir + routesPath)
	if err != nil {
		log.Fatal(err)
	}
	return routes
}

// CSVError represents errors that occur during CSV parsing
type CSVError struct {
	FilePath string
	Message  string
}

func (e CSVError) Error() string {
	return fmt.Sprintf("CSV error in %s: %s", e.FilePath, e.Message)
}

// parseCSV parses each row in a CSV file into a struct of type R.
// Each field in R to be parsed must specify a csv tag denoting the column header.
// CSVError is returned if file could not be parsed.
func parseCSV[R any](filePath string) ([]R, error) {
	var zero R
	r := reflect.TypeOf(zero)
	rows := []R{}

	// Only accept structs
	if r.Kind() != reflect.Struct {
		return nil, CSVError{
			FilePath: filePath,
			Message:  "parseCSV must be passed a struct for R",
		}
	}

	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, CSVError{
			FilePath: filePath,
			Message:  fmt.Sprintf("parseCSV failed to read file: %v", err),
		}
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
