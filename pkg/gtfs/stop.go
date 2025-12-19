package gtfs

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type Stop struct {
	StopId        string
	StopName      string
	StopLat       float64
	StopLon       float64
	LocationType  string
	ParentStation string
	RouteIds      map[string]bool
}

type Shape struct {
	ShapeId    string
	ShapePtSeq int
	ShapePtLat float64
	ShapePtLon float64
}

type Route struct {
	RouteId        string
	AgencyId       string
	RouteShortName string
	RouteLongName  string
	RouteDesc      string
	RouteType      string
	RouteUrl       string
	RouteColor     string
	RouteTextColor string
	RouteSortOrder int
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
		stopCoords := [2]float64{stop.StopLat, stop.StopLon}
		for _, shape := range shapes {
			shapeCoords := [2]float64{shape.ShapePtLat, shape.ShapePtLon}
			if stopCoords == shapeCoords {
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
		if stop.LocationType == "1" {
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
	var stops []Stop

	bytes, err := os.ReadFile(dataDir + stopsPath)
	if err != nil {
		log.Fatal(err)
	}

	content := string(bytes)
	rows := strings.Split(content, "\n")

	for _, row := range rows {
		fields := strings.Split(row, ",")

		if len(fields) < 6 {
			continue
		}

		StopId := fields[0]
		StopName := fields[1]
		StopLat, _ := strconv.ParseFloat(fields[2], 64)
		StopLon, _ := strconv.ParseFloat(fields[3], 64)
		LocationType := fields[4]
		ParentStation := fields[5]

		stop := Stop{
			StopId:        StopId,
			StopName:      StopName,
			StopLat:       StopLat,
			StopLon:       StopLon,
			LocationType:  LocationType,
			ParentStation: ParentStation,
		}

		stops = append(stops, stop)
	}

	return stops
}

func getShapes() []Shape {
	var shapes []Shape

	bytes, err := os.ReadFile(dataDir + shapesPath)
	if err != nil {
		log.Fatal(err)
	}

	content := string(bytes)
	rows := strings.Split(content, "\n")

	for _, row := range rows {
		fields := strings.Split(row, ",")

		if len(fields) < 4 {
			continue
		}

		ShapeId := fields[0]
		ShapePtSeq, _ := strconv.Atoi(fields[1])
		ShapePtLat, _ := strconv.ParseFloat(fields[2], 64)
		ShapePtLon, _ := strconv.ParseFloat(fields[3], 64)

		shape := Shape{
			ShapeId:    ShapeId,
			ShapePtSeq: ShapePtSeq,
			ShapePtLat: ShapePtLat,
			ShapePtLon: ShapePtLon,
		}

		shapes = append(shapes, shape)
	}

	return shapes
}

func GetRoutes() []Route {
	var routes []Route

	bytes, err := os.ReadFile(dataDir + routesPath)
	if err != nil {
		log.Fatal(err)
	}

	content := string(bytes)
	rows := strings.Split(content, "\n")

	for _, row := range rows {
		chars := strings.Split(row, "")

		fields := []string{}
		start := 0
		inQuotes := false
		for i, char := range chars {
			if char == "\"" {
				inQuotes = !inQuotes
			} else if !inQuotes && (char == "," || i == len(chars)-1) {
				field := strings.Join(chars[start:i], "")
				fields = append(fields, field)
				start = i + 1
			}
		}

		if len(fields) < 9 {
			continue
		}

		RouteId := fields[0]
		AgencyId := fields[1]
		RouteShortName := fields[2]
		RouteLongName := fields[3]
		RouteDesc := fields[4]
		RouteType := fields[5]
		RouteUrl := fields[6]
		RouteColor := "#" + fields[7]
		RouteTextColor := "#" + fields[8]
		RouteSortOrder, _ := strconv.Atoi(fields[9])

		route := Route{
			RouteId:        RouteId,
			AgencyId:       AgencyId,
			RouteShortName: RouteShortName,
			RouteLongName:  RouteLongName,
			RouteDesc:      RouteDesc,
			RouteType:      RouteType,
			RouteUrl:       RouteUrl,
			RouteColor:     RouteColor,
			RouteTextColor: RouteTextColor,
			RouteSortOrder: RouteSortOrder,
		}

		routes = append(routes, route)
	}

	fmt.Println(routes)

	return routes
}
