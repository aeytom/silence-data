package main

import (
	"context"
	"fmt"
	"log"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

func main() {
	ParseArgs()

	ixclient := influxdb2.NewClient(Conf.Influx.Url, Conf.Influx.Token)

	silence := Silence{}
	if err := silence.Login(Conf.Silence.Email, Conf.Silence.Password); err != nil {
		log.Fatalln(err)
	}

	var err error
	var profile ProfileResponse
	if profile, err = silence.Me(); err != nil {
		log.Fatalln(err)
	} else {
		log.Printf("%#v", profile)
	}

	// silence.TripsList(profile.Id, 100)
	// panic("Schluss")

	writeAPI := ixclient.WriteAPIBlocking(Conf.Influx.Org, Conf.Influx.Bucket)
	for {
		scooters, err := silence.Details()
		if err != nil {
			log.Fatalln(err)
		}
		log.Printf("%#v", scooters)

		for s := 0; s < len(scooters); s++ {
			c := scooters[s]
			tags := map[string]string{
				"id":       c.Id,
				"model":    c.Model,
				"revision": c.Revision,
				"color":    c.Color,
				"name":     c.Name,
				"imei":     c.Imei,
				"frameno":  c.FrameNo,
				"firmware": c.TrackingDevice.FirmwareVersion,
				"battery":  fmt.Sprint(c.BatteryId),
			}
			fields := map[string]interface{}{
				"speed":    c.LastLocation.CurrentSpeed,
				"bsoc":     c.BatterySoc,
				"btemp":    c.BatteryTemperature,
				"odo":      c.Odometer,
				"mtemp":    c.MotorTemperature,
				"itemp":    c.InverterTemperature,
				"range":    c.Range,
				"velocity": c.Velocity,
				"lat":      c.LastLocation.Latitude,
				"lon":      c.LastLocation.Longitude,
			}
			lt := c.LastConnection
			if c.LastLocation.Time != "" {
				lt = c.LastLocation.Time
			} else if c.LastReportTime != "" {
				lt = c.LastReportTime
			}
			log.Printf("%#v", lt)
			if lrt, err := time.Parse(time.RFC3339, lt); err != nil {
				log.Fatalln(err)
			} else {
				point := write.NewPoint("scooter", tags, fields, lrt)
				if err := writeAPI.WritePoint(context.Background(), point); err != nil {
					log.Fatal(err)
				}
			}
		}
		time.Sleep(30 * time.Second)
	}

}
