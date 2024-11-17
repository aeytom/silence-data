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
			Connections:      map[string]string{"imei": scooter.Imei, "btMac": scooter.BtMac},
			HwVersion:        scooter.Revision,
			Identifiers:      []string{scooter.Id},
			Manufacturer:     "Scutum",
			Model:            scooter.Model,
			ModelId:          "",
			Name:             scooter.Name,
			SwVersion:        scooter.TrackingDevice.FirmwareVersion,
		},
		Origin: Origin{
			Name:      scooter.Name,
			SwVersion: scooter.Revision,
		},
		Components: map[string]DiscoveryPayload{
			"LastReportTime": {
				DeviceClass:   "timestamp",
				Name:          "LastReportTime",
				ValueTemplate: "{{ value_json.lastReport_time }}",
			},
			"LastLocationTime": {
				DeviceClass:   "timestamp",
				Name:          "LastLocationTime",
				ValueTemplate: "{{ value_json.lastLocation.time }}",
			},
			"LastConnectionTime": {
				DeviceClass:   "timestamp",
				Name:          "LastConnectionTime",
				ValueTemplate: "{{ value_json.lastConnection }}",
			},
			"BatterySoc": {
				DeviceClass:       "battery",
				Name:              "BatterySoc",
				StateClass:        "measurement",
				UnitOfMeasurement: "%",
				ValueTemplate:     "{{ value_json.batterySoc }}",
			},
			"BatteryTemperature": {
				DeviceClass:       "temperature",
				Name:              "BatteryTemperature",
				StateClass:        "measurement",
				UnitOfMeasurement: "°C",
				ValueTemplate:     "{{ value_json.batteryTemperature }}",
			},
			"MotorTemperature": {
				DeviceClass:       "temperature",
				Name:              "MotorTemperature",
				StateClass:        "measurement",
				UnitOfMeasurement: "°C",
				ValueTemplate:     "{{ value_json.motorTemperature }}",
			},
			"InverterTemperature": {
				DeviceClass:       "temperature",
				Name:              "InverterTemperature",
				StateClass:        "measurement",
				UnitOfMeasurement: "°C",
				ValueTemplate:     "{{ value_json.inverterTemperature }}",
			},
			"Odometer": {
				DeviceClass:       "distance",
				Name:              "Odometer",
				StateClass:        "total_increasing",
				UnitOfMeasurement: "km",
				ValueTemplate:     "{{ value_json.odometer }}",
			},
			"Range": {
				DeviceClass:       "distance",
				Name:              "Range",
				StateClass:        "measurement",
				UnitOfMeasurement: "km",
				ValueTemplate:     "{{ value_json.range }}",
			},
			"Velocity": {
				DeviceClass:       "speed",
				Name:              "Velocity",
				StateClass:        "measurement",
				UnitOfMeasurement: "km/h",
				ValueTemplate:     "{{ value_json.velocity }}",
			},
			"LastLocation": {
				DeviceClass:            "device_tracker",
				Name:                   "LastLocation",
				JsonAttributesTopic:    fmt.Sprintf(LocationTemplate, scooter.Id),
				JsonAttributesTemplate: "{{ value_json }}",
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
