package hass

import (
	"fmt"

	"github.com/aeytom/silence-data/silence"
)

const (
	AvailabilityTemplate = "silence/%s/availability"
	LocationTemplate     = "silence/%s/location"
	StateTemplate        = "silence/%s/scooter/state"
)

func RegisterScooter(c *Client, scooter silence.ScooterResp) {
	c.SendDiscovery(scooter)
	c.SendAvailability(scooter, true)
	c.Scooters = append(c.Scooters, &scooter)
}

func (c *Client) SendDiscovery(scooter silence.ScooterResp) {
	dev := DeviceDiscovery{
		StateTopic: fmt.Sprintf(StateTemplate, scooter.Id),
		Availability: Availability{
			Topic: fmt.Sprintf(AvailabilityTemplate, scooter.Id),
		},
		Device: HaDevice{
			ConfigurationUrl: "",
			Connections: [][]string{
				{"imei", scooter.Imei},
				{"btMac", scooter.BtMac}},
			HwVersion:    scooter.Revision,
			Identifiers:  []string{scooter.Id},
			Manufacturer: "Scutum",
			Model:        scooter.Model,
			ModelId:      "",
			Name:         scooter.Name,
			SwVersion:    scooter.TrackingDevice.FirmwareVersion,
		},
		Origin: Origin{
			Name:      scooter.Name,
			SwVersion: scooter.Revision,
		},
		Components: map[string]DiscoveryPayload{
			"LastReportTime": {
				Platform:      "sensor",
				DeviceClass:   "timestamp",
				Name:          "LastReportTime",
				UniqueId:      scooter.Id + "-LastReportTime",
				ValueTemplate: "{{ value_json.lastReportTime }}",
			},
			"LastLocationTime": {
				Platform:      "sensor",
				DeviceClass:   "timestamp",
				Name:          "LastLocationTime",
				UniqueId:      scooter.Id + "-LastLocationTime",
				ValueTemplate: "{{ value_json.lastLocation.time }}",
			},
			"LastConnectionTime": {
				Platform:      "sensor",
				DeviceClass:   "timestamp",
				Name:          "LastConnectionTime",
				UniqueId:      scooter.Id + "-LastConnectionTime",
				ValueTemplate: "{{ value_json.lastConnection }}",
			},
			"BatterySoc": {
				Platform:          "sensor",
				DeviceClass:       "battery",
				Name:              "BatterySoc",
				StateClass:        "measurement",
				UnitOfMeasurement: "%",
				UniqueId:          scooter.Id + "-BatterySoc",
				ValueTemplate:     "{{ value_json.batterySoc }}",
			},
			"BatteryTemperature": {
				Platform:          "sensor",
				DeviceClass:       "temperature",
				Name:              "BatteryTemperature",
				StateClass:        "measurement",
				UnitOfMeasurement: "°C",
				UniqueId:          scooter.Id + "-BatteryTemperature",
				ValueTemplate:     "{{ value_json.batteryTemperature }}",
			},
			"MotorTemperature": {
				Platform:          "sensor",
				DeviceClass:       "temperature",
				Name:              "MotorTemperature",
				StateClass:        "measurement",
				UnitOfMeasurement: "°C",
				UniqueId:          scooter.Id + "-MotorTemperature",
				ValueTemplate:     "{{ value_json.motorTemperature }}",
			},
			"InverterTemperature": {
				Platform:          "sensor",
				DeviceClass:       "temperature",
				Name:              "InverterTemperature",
				StateClass:        "measurement",
				UnitOfMeasurement: "°C",
				UniqueId:          scooter.Id + "-InverterTemperature",
				ValueTemplate:     "{{ value_json.inverterTemperature }}",
			},
			"Odometer": {
				Platform:          "sensor",
				DeviceClass:       "distance",
				Name:              "Odometer",
				StateClass:        "total_increasing",
				UnitOfMeasurement: "km",
				UniqueId:          scooter.Id + "-Odometer",
				ValueTemplate:     "{{ value_json.odometer }}",
			},
			"Range": {
				Platform:          "sensor",
				DeviceClass:       "distance",
				Name:              "Range",
				StateClass:        "measurement",
				UnitOfMeasurement: "km",
				UniqueId:          scooter.Id + "-Range",
				ValueTemplate:     "{{ value_json.range }}",
			},
			"Velocity": {
				Platform:          "sensor",
				DeviceClass:       "speed",
				Name:              "Velocity",
				StateClass:        "measurement",
				UnitOfMeasurement: "km/h",
				UniqueId:          scooter.Id + "-Velocity",
				ValueTemplate:     "{{ value_json.velocity }}",
			},
			"LastLocation": {
				Platform:            "device_tracker",
				Name:                "LastLocation",
				JsonAttributesTopic: fmt.Sprintf(LocationTemplate, scooter.Id),
				UniqueId:            scooter.Id + "-LastLocation",
			},
		},
	}

	topic := fmt.Sprintf("%s/%s/%s/config", c.DiscoveryPrefix, "device", scooter.Id)
	c.Send(topic, 0, true, dev)
}

func (c *Client) Disconnect() {
	for _, scooter := range c.Scooters {
		c.SendAvailability(*scooter, false)
	}
	c.Client.Disconnect(250)
}

func SendStatus(c *Client, scooter silence.ScooterResp) {
	c.SendAvailability(scooter, true)
	c.Send(fmt.Sprintf(StateTemplate, scooter.Id), 0, false, scooter)
}

func SendLocation(c *Client, scooter silence.ScooterResp) {
	ja := DeviceTrackerAttributes{
		Longitude: scooter.LastLocation.Longitude,
		Latitude:  scooter.LastLocation.Latitude,
	}
	c.Send(fmt.Sprintf(LocationTemplate, scooter.Id), 0, false, ja)
}

func (c *Client) SendAvailability(scooter silence.ScooterResp, available bool) {
	pl := "offline"
	if available {
		pl = "online"
	}
	c.Send(fmt.Sprintf(AvailabilityTemplate, scooter.Id), 0, !available, pl)
}
