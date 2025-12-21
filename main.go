package main

import (
	"log"
	"nyct-feed/pkg/gtfs"
)

var stopIds = []string{
	"A46N",
	"A46S",
	"239N",
	"239S",
}

func main() {
	schedule := gtfs.GetSchedule()

	log.Println(schedule.Stops[:10])
	log.Println(schedule.StopTimes[:10])
	log.Println(schedule.Trips[:10])
	log.Println(schedule.Routes[:10])
	log.Println(schedule.Shapes[:10])

	// bytes, _ := os.ReadFile("data/stops.txt")
	// rows, _ := gtfs.ParseCSV(bytes, gtfs.Stop{})
	// for _, row := range rows[:10] {
	// 	fmt.Println(row)
	// }
}
