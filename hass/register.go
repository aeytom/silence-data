package hass

import (
	"fmt"
	"strings"

	"github.com/aeytom/silence-data/silence"
)

const (
	AvailabilityTemplate = "silence/%s/availability"
	LocationTemplate     = "silence/%s/location"
	StateTemplate        = "silence/%s/scooter/state"
)

func RegisterScooter(c *Client, scooter *silence.ScooterResp) {
	dev := DiscoverDevice{
		ConfigurationUrl: "",
		HwVersion:        scooter.Revision,
		Identifiers:      []string{scooter.Id},
		Manufacturer:     "Scutum",
		Model:            scooter.Model,
		ModelId:          "",
		Name:             scooter.Name,
		SwVersion:        scooter.TrackingDevice.FirmwareVersion,
	}

	registerDevice(scooter, dev, c)

	registerMeasurement(scooter, dev, c, "BatterySoc", "%", "battery", "{{ value_json.batterySoc }}")
	registerMeasurement(scooter, dev, c, "BatteryTemperature", "°C", "temperature", "{{ value_json.batteryTemperature }}")
	registerMeasurement(scooter, dev, c, "MotorTemperature", "°C", "temperature", "{{ value_json.motorTemperature }}")
	registerMeasurement(scooter, dev, c, "InverterTemperature", "°C", "temperature", "{{ value_json.inverterTemperature }}")
	registerMeasurement(scooter, dev, c, "Range", "km", "distance", "{{ value_json.range }}")
	registerMeasurement(scooter, dev, c, "Velocity", "km/h", "speed", "{{ value_json.velocity }}")

	registerTracker(scooter, dev, c, "LastLocation")
	SendAvailability(c, *scooter, true)
	c.handles = append(c.handles, scooter)
}

func registerMeasurement(scooter *silence.ScooterResp, dev DiscoverDevice, c *Client, name string, unit_of_measurement string, device_class string, template string) {
	oid := strings.ToLower(scooter.Id + "-" + name)
	if strings.Contains(oid, "battery") {
		oid += fmt.Sprintf("-%d", scooter.BatteryId)
	}
	p := DiscoveryPayload{
		Name:              scooter.Name + " " + name,
		StateClass:        "measurement",
		StateTopic:        fmt.Sprintf(StateTemplate, scooter.Id),
		AvailabilityTopic: fmt.Sprintf(AvailabilityTemplate, scooter.Id),
		UnitOfMeasurement: unit_of_measurement,
		ObjectId:          oid,
		DeviceClass:       device_class,
		ValueTemplate:     template,
		Device:            dev,
	}
	topic := fmt.Sprintf("%s/%s/%s/config", c.DiscoveryPrefix, "sensor", oid)
	c.Send(topic, 0, true, p)
}

func registerDevice(scooter *silence.ScooterResp, dev DiscoverDevice, c *Client) {
	p := DiscoveryPayload{
		Name:              scooter.Name,
		AvailabilityTopic: fmt.Sprintf(AvailabilityTemplate, scooter.Id),
		StateTopic:        fmt.Sprintf(StateTemplate, scooter.Id),
		ObjectId:          scooter.Id,
		Device:            dev,
	}
	topic := fmt.Sprintf("%s/%s/%s/config", c.DiscoveryPrefix, "device", scooter.Id)
	c.Send(topic, 0, true, p)
}

func registerTracker(scooter *silence.ScooterResp, dev DiscoverDevice, c *Client, name string) {
	oid := strings.ToLower(scooter.Id + "-" + name)
	p := DiscoveryPayload{
		Name:                   scooter.Name + " " + name,
		AvailabilityTopic:      fmt.Sprintf(AvailabilityTemplate, scooter.Id),
		JsonAttributesTopic:    fmt.Sprintf(LocationTemplate, scooter.Id),
		JsonAttributesTemplate: "{{ value_json }}",
		ObjectId:               oid,
		Device:                 dev,
	}
	topic := fmt.Sprintf("%s/%s/%s/config", c.DiscoveryPrefix, "sensor", oid)
	c.Send(topic, 0, true, p)
}

func (c *Client) Disconnect() {
	for _, scooter := range c.handles {
		SendAvailability(c, *scooter, false)
	}
	c.Client.Disconnect(250)
}

func SendStatus(c *Client, scooter *silence.ScooterResp) {
	SendAvailability(c, *scooter, true)
	c.Send(fmt.Sprintf(StateTemplate, scooter.Id), 0, false, scooter)
}

func SendLocation(c *Client, scooter silence.ScooterResp) {
	scooter.LastLocation.Altitude = 0
	scooter.LastLocation.CurrentSpeed = 0
	scooter.LastLocation.Time = ""
	c.Send(fmt.Sprintf(LocationTemplate, scooter.Id), 0, false, scooter.LastLocation)
}

func SendAvailability(c *Client, scooter silence.ScooterResp, available bool) {
	pl := "offline"
	if available {
		pl = "online"
	}
	c.Send(fmt.Sprintf(AvailabilityTemplate, scooter.Id), 0, !available, pl)
}
