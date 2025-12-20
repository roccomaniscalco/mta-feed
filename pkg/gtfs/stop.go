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
	return parseCSV[Stop](dataDir + stopsPath)
}

func getShapes() []Shape {
	return parseCSV[Shape](dataDir + shapesPath)
}

func GetRoutes() []Route {
	return parseCSV[Route](dataDir + routesPath)
}

func parseCSV[T any](filePath string) []T {
	// Create a zero value to get type information
	var zero T
	t := reflect.TypeOf(zero)
	structs := []T{}

	// Only accept structs
	if t.Kind() != reflect.Struct {
		log.Fatal("parseCsv only works with struct types")
	}

	bytes, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal(fmt.Sprintf("Error reading file %s: %s", filePath, err))
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
		newValue := reflect.New(t).Elem()

		// Populate fields based on CSV cells
		for i := 0; i < t.NumField(); i++ {
			field := newValue.Field(i)
			fieldType := t.Field(i)

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

		structs = append(structs, newValue.Interface().(T))
	}

	return structs
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
