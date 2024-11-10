package hass

import (
	"bytes"
	"encoding/json"
	"log"
	"os"

	"github.com/aeytom/silence-data/silence"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	DiscoveryPrefix = "homeassistant"
)

type Config struct {
	MqttServer      string `yaml:"mqtt_server,omitempty" json:"mqtt_server,omitempty"`
	MqttClientId    string `yaml:"mqtt_client_id,omitempty" json:"mqtt_client_id,omitempty"`
	MqttUser        string `yaml:"mqtt_user,omitempty" json:"mqtt_user,omitempty"`
	MqttPassword    string `yaml:"mqtt_password,omitempty" json:"mqtt_password,omitempty"`
	DiscoveryPrefix string `yaml:"discovery_prefix,omitempty" json:"discovery_prefix,omitempty"`
}

type Meter interface {
	HassRegister(c *Client) *Handle
	HassSendValue()
}

type MeterState struct {
	Value             float32 `json:"value,omitempty"`
	UnitOfMeasurement string  `json:"unit_of_measurement,omitempty"`
}

type DiscoverDevice struct {
	ConfigurationUrl string   `json:"configuration_url,omitempty"`
	HwVersion        string   `json:"hw_version,omitempty"`
	Identifiers      []string `json:"identifiers,omitempty"`
	Manufacturer     string   `json:"manufacturer,omitempty"`
	Model            string   `json:"model,omitempty"`
	ModelId          string   `json:"model_id,omitempty"`
	Name             string   `json:"name,omitempty"`
	SwVersion        string   `json:"sw_version,omitempty"`
}

type DiscoveryPayload struct {
	AvailabilityTopic      string         `json:"availability_topic,omitempty"`
	CommandTopic           string         `json:"command_topic,omitempty"`
	Device                 DiscoverDevice `json:"device,omitempty"`
	DeviceClass            string         `json:"device_class,omitempty"`
	Name                   string         `json:"name,omitempty"`
	ObjectId               string         `json:"object_id,omitempty"`
	StateClass             string         `json:"state_class,omitempty"`
	StateTopic             string         `json:"state_topic,omitempty"`
	JsonAttributesTopic    string         `json:"json_attributes_topic,omitempty"`
	JsonAttributesTemplate string         `json:"json_attributes_template,omitempty"`
	SupportUrl             string         `json:"support_url,omitempty"`
	SwVersion              string         `json:"sw_version,omitempty"`
	UniqueId               string         `json:"unique_id,omitempty"`
	UnitOfMeasurement      string         `json:"unit_of_measurement,omitempty"`
	ValueTemplate          string         `json:"value_template,omitempty"`
}

type Client struct {
	Client          mqtt.Client
	DiscoveryPrefix string
	handles         []*silence.ScooterResp
}

type Handle struct {
	Mqtt   *Client
	Object DiscoveryPayload
}

func Connect(cfg Config) *Client {
	c := Client{}
	opts := mqtt.NewClientOptions().AddBroker(cfg.MqttServer).SetClientID(cfg.MqttClientId)
	opts.SetUsername(cfg.MqttUser)
	opts.SetPassword(cfg.MqttPassword)
	opts.SetAutoReconnect(true)
	//
	mqtt.ERROR = log.New(os.Stdout, "[ERROR] ", 0)
	mqtt.CRITICAL = log.New(os.Stdout, "[CRIT] ", 0)
	mqtt.WARN = log.New(os.Stdout, "[WARN]  ", 0)
	// mqtt.DEBUG = log.New(os.Stdout, "[DEBUG] ", 0)
	//
	c.Client = mqtt.NewClient(opts)
	if token := c.Client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	if cfg.DiscoveryPrefix == "" {
		c.DiscoveryPrefix = DiscoveryPrefix
	} else {
		c.DiscoveryPrefix = cfg.DiscoveryPrefix
	}

	return &c
}

func (c *Client) Send(topic string, qos byte, retain bool, payload interface{}) {
	switch p := payload.(type) {
	case string:
		c.sendBytes(topic, qos, retain, []byte(p))
	case []byte:
		c.sendBytes(topic, qos, retain, p)
	case bytes.Buffer:
		c.sendBytes(topic, qos, retain, p.Bytes())
	default:
		if msg, err := json.Marshal(p); err == nil {
			c.sendBytes(topic, qos, retain, msg)
		} else {
			mqtt.ERROR.Println(err)
		}
	}
}

func (c *Client) sendBytes(topic string, qos byte, retain bool, msg []byte) {
	mqtt.DEBUG.Printf("topic: %s message: %s", topic, msg)
	t := c.Client.Publish(topic, qos, retain, msg)
	go func() {
		_ = t.Wait()
		if t.Error() != nil {
			mqtt.ERROR.Println(t.Error())
		}
	}()
}
