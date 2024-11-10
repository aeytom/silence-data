package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aeytom/silence-data/hass"
	"github.com/aeytom/silence-data/silence"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	ilog "github.com/influxdata/influxdb-client-go/v2/log"
)

func main() {
	ParseArgs()

	ha := hass.Connect(Conf.HomeAssistant)
	defer ha.Disconnect()

	ixclient := influxdb2.NewClient(Conf.Influx.Url, Conf.Influx.Token)
	ixclient.Options().SetLogLevel(ilog.DebugLevel)

	si := silence.Silence{}
	if err := si.Login(Conf.Silence.Email, Conf.Silence.Password); err != nil {
		log.Fatalln(err)
	}

	var err error
	var profile silence.ProfileResponse
	if profile, err = si.Me(); err != nil {
		log.Fatalln(err)
	} else {
		log.Printf("%#v", profile)
	}

	// silence.TripsList(profile.Id, 100)
	// panic("Schluss")

	scooters, err := si.Details()
	if err != nil {
		log.Fatalln(err)
	} else {
		for _, sc := range scooters {
			hass.RegisterScooter(ha, &sc)
		}
	}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(30 * time.Second)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case s := <-sigs:
				log.Print("Got signal: ", s)
				done <- true
				return
			case t := <-ticker.C:
				log.Println("Tick at", t)

				scooters, err := si.Details()
				if err != nil {
					log.Fatalln(err)
				}
				log.Printf("%#v", scooters)

				for _, scooter := range scooters {
					hass.SendStatus(ha, &scooter)
					hass.SendLocation(ha, scooter)
					sendToInflux(ixclient, scooter)
				}
			}
		}
	}()

	<-done
	ticker.Stop()
}

func sendToInflux(ixclient influxdb2.Client, scooter silence.ScooterResp) {
	writeAPI := ixclient.WriteAPIBlocking(Conf.Influx.Org, Conf.Influx.Bucket)
	tags := map[string]string{
		"id":       scooter.Id,
		"model":    scooter.Model,
		"revision": scooter.Revision,
		"color":    scooter.Color,
		"name":     scooter.Name,
		"imei":     scooter.Imei,
		"frameno":  scooter.FrameNo,
		"firmware": scooter.TrackingDevice.FirmwareVersion,
		"battery":  fmt.Sprint(scooter.BatteryId),
	}
	fields := map[string]interface{}{
		"speed":    scooter.LastLocation.CurrentSpeed,
		"bsoc":     scooter.BatterySoc,
		"btemp":    scooter.BatteryTemperature,
		"odo":      scooter.Odometer,
		"mtemp":    scooter.MotorTemperature,
		"itemp":    scooter.InverterTemperature,
		"range":    scooter.Range,
		"velocity": scooter.Velocity,
		"lat":      scooter.LastLocation.Latitude,
		"lon":      scooter.LastLocation.Longitude,
	}
	lt := scooter.LastConnection
	if scooter.LastLocation.Time != "" {
		lt = scooter.LastLocation.Time
	} else if scooter.LastReportTime != "" {
		lt = scooter.LastReportTime
	}
	log.Printf("last location/report time %#v", lt)
	if lrt, err := time.Parse(time.RFC3339, lt); err != nil {
		log.Fatalln(err)
	} else {
		point := influxdb2.NewPoint("scooter", tags, fields, lrt)
		if err := writeAPI.WritePoint(context.Background(), point); err != nil {
			log.Fatal(err)
		}
	}
}
