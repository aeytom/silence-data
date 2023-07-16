package main

import (
	"flag"
	"log"
	"os"

	"github.com/go-yaml/yaml"
)

type DotEnv struct {
	Silence struct {
		Email    string `yaml:"email" json:"email"`
		Password string `yaml:"password" json:"password"`
	} `yaml:"silence" json:"silence"`
	Influx struct {
		Org    string `yaml:"org,omitempty" json:"org,omitempty"`
		Bucket string `yaml:"bucket,omitempty" json:"bucket,omitempty"`
		Token  string `yaml:"token" json:"token"`
		Url    string `yaml:"url" json:"url"`
	} `yaml:"influx" json:"influx"`
}

// Command line args
var (
	Conf DotEnv
)

// ParseArgs parses command line flags
func ParseArgs() {

	envPath := getEnvArg("DOT_ENV", "dotEnv", ".env.yaml", "dot env path")
	flag.Parse()

	ed, err := os.ReadFile(*envPath)
	if err != nil {
		log.Fatalln(err)
	}

	err = yaml.Unmarshal([]byte(ed), &Conf)
	if err != nil {
		log.Fatalln(err)
	}

	if Conf.Silence.Email == "" {
		log.Fatal("You must set .env.yaml silence.email")
	}
	if Conf.Silence.Password == "" {
		log.Fatal("You must set .env.yaml silence.password")
	}

	if Conf.Influx.Token == "" {
		log.Fatal("You must set .env.yaml influx.token")
	}
	if Conf.Influx.Bucket == "" {
		Conf.Influx.Bucket = "silence"
	}
	if Conf.Influx.Org == "" {
		Conf.Influx.Org = "primary"
	}
	if Conf.Influx.Url == "" {
		log.Fatal("You must set .env.yaml influx.url")
	}

	log.Printf("%#v", Conf)
}

func getEnvArg(env string, arg string, dflt string, usage string) *string {
	ev, avail := os.LookupEnv(env)
	if avail {
		dflt = ev
	}
	v := flag.String(arg, dflt, usage)
	return v
}
