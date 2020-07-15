package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"time"
)

type ConfigFile struct {
	VictoriaMetrics Server
	Nomad           []Server
}

type Server struct {
	URL  string
	Port string
}

func loadConfig(path string) ([]string, string, time.Duration) {
	var metricsAddress string
	var nomadAddresses []string

	data, err := ioutil.ReadFile(path)
	var config ConfigFile
	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Fatal(err)
	}

	metricsAddress = config.VictoriaMetrics.URL + ":" + config.VictoriaMetrics.Port

	for _, server := range config.Nomad {
		nomadAddresses = append(nomadAddresses, server.URL+":"+server.Port)
	}

	duration, err := time.ParseDuration("1m")
	if err != nil {
		log.Fatal(err)
	}

	return nomadAddresses, metricsAddress, duration
}
